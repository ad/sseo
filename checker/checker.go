package checker

import (
	"context"
	"fmt"
	"time"

	"github.com/ad/lru"
	"github.com/ad/sseo/rules"
	"github.com/alitto/pond"
)

type Checker struct {
	LRU *lru.Cache[string, Result]
	// Results chan Result
	Pool  *pond.WorkerPool
	Tasks chan Task
}

type Result struct {
	URL       string
	Status    string
	Error     error
	CreatedAt time.Time
	Checks    []string
}

type Task struct {
	URL   string
	Force bool
}

func InitChecker() *Checker {
	// Create a worker pool
	pool := pond.New(
		3,    // Number of workers
		1000, // Size of the queue
		pond.MinWorkers(1),
		pond.IdleTimeout(1*time.Second),
		pond.PanicHandler(panicHandler),
		pond.Strategy(pond.Lazy()),
	)
	// defer pool.StopAndWait()

	checker := &Checker{
		LRU: lru.New[string, Result](lru.WithCapacity(1000)),
		// Results: make(chan Result),
		Tasks: make(chan Task),
		Pool:  pool,
	}

	checker.Start()

	return checker
}

func panicHandler(r interface{}) {
	fmt.Printf("panic: %v\n", r)
}

func (c *Checker) Start() {
	// go func() {
	// 	for r := range c.Results {
	// 		c.LRU.Set(r.URL, r)
	// 	}
	// }()

	go func() {
		defer c.Pool.StopAndWait()

		// Create a task group associated to a context
		group, _ := c.Pool.GroupContext(context.Background())

		for t := range c.Tasks {
			cachedResult, ok := c.LRU.Get(t.URL)

			if ok {
				if cachedResult.Status == "processing" {
					fmt.Printf("PROCESSING: %s %s, Error: %v, CreatedAt: %v, Checks: %v\n", cachedResult.Status, cachedResult.URL, cachedResult.Error, cachedResult.CreatedAt, cachedResult.Checks)

					continue
				}
			}

			if !t.Force {
				if ok {
					fmt.Printf("CACHED: %s %s, Error: %v, CreatedAt: %v, Checks: %v\n", cachedResult.Status, cachedResult.URL, cachedResult.Error, cachedResult.CreatedAt, cachedResult.Checks)
					// c.Results <- cachedResult

					// if cachedResult.Status == "processing" {
					continue
					// }
				}
			}

			r := Result{
				URL:       t.URL,
				Status:    "processing",
				CreatedAt: time.Now(),
			}

			c.LRU.Set(t.URL, r)

			group.Submit(func() error {
				fmt.Printf("processing: %s %t\n", t.URL, t.Force)
				time.Sleep(1 * time.Second)

				result, err := checkURL(t.URL)

				fmt.Printf("FETCHED: %s, Error: %v, CreatedAt: %v, Checks: %v\n", t.URL, err, time.Now(), result)

				c.LRU.Set(r.URL, Result{
					URL:       t.URL,
					Status:    "done",
					Error:     err,
					CreatedAt: time.Now(),
					Checks:    result,
				})

				// c.Results <- Result{
				// 	URL:       t.URL,
				// 	Status:    "done",
				// 	Error:     err,
				// 	CreatedAt: time.Now(),
				// 	Checks:    result,
				// }

				return nil
			})
		}

		err := group.Wait()
		if err != nil {
			fmt.Printf("Failed to fetch URLs: %v\n", err)
		}
	}()
}

func checkURL(url string) ([]string, error) {
	// fmt.Println("start testing... ", url)

	ruleChecker, errRules := rules.NewRulesWith(url)
	if errRules != nil {
		// fmt.Println("failed to create rules:", errRules)

		return []string{}, errRules
	}
	ruleChecker.AddRule(rules.WithStatus(ruleChecker.StatusCode, []int{200}))
	ruleChecker.AddRule(rules.WithTitle(ruleChecker.Parsed))
	ruleChecker.AddRule(rules.WithDescription(ruleChecker.Parsed))
	ruleChecker.AddRule(rules.WithHeading(ruleChecker.Parsed))
	ruleChecker.AddRule(rules.WithRobotsTXT(ruleChecker.URL))
	ruleChecker.AddRule(rules.WithSitemap(ruleChecker.URL))

	errors := ruleChecker.Check()

	if len(errors) > 0 {
		// fmt.Println(url, "errors:", strings.Join(errors, ", "))
		return errors, nil
	}

	return []string{"ok"}, errRules
}
