package main

import (
	"fmt"
	"strings"

	"github.com/ad/sseo/rules"
	"github.com/sourcegraph/conc/pool"
)

/*
 наличие и заполнение <title> (длина текста)
 наличие и заполнение <meta name="description" content=""> (длина текста)
 Наличие H1
 Приоритет заголовков H1-H4
 Наличие robots.txt и Allow в нем
 Наличие sitemap.xml
*/

func main() {
	fmt.Println("start testing...")

	urls := []string{
		"https://google.com",
		"https://yandex.ru",
		"https://mail.ru",
		"https://rambler.ru",
		"https://yahoo.com",
		"https://bing.com",
		"https://duckduckgo.com",
		"https://ask.com",
		"https://wow.com",
		"https://excite.com",
		"https://alhea.com",
		"https://info.com",
	}

	p := pool.New().WithMaxGoroutines(2)
	for _, elem := range urls {
		elem := elem
		p.Go(func() {
			checkURL(elem)
		})
	}
	p.Wait()

	fmt.Println("end testing...")
	fmt.Println("")
}

func checkURL(url string) {
	// fmt.Println("start testing... ", url)

	ruleChecker, errRules := rules.NewRulesWith(url)
	if errRules != nil {
		fmt.Println("failed to create rules:", errRules)

		return
	}
	ruleChecker.AddRule(rules.WithStatus(ruleChecker.StatusCode, []int{200}))
	ruleChecker.AddRule(rules.WithTitle(ruleChecker.Parsed))
	ruleChecker.AddRule(rules.WithDescription(ruleChecker.Parsed))
	ruleChecker.AddRule(rules.WithHeading(ruleChecker.Parsed))
	ruleChecker.AddRule(rules.WithRobotsTXT(ruleChecker.URL))
	ruleChecker.AddRule(rules.WithSitemap(ruleChecker.URL))

	errors := ruleChecker.Check()

	if len(errors) > 0 {
		fmt.Println(url, "errors:", strings.Join(errors, ", "))
	} else {
		fmt.Println(url, "OK")
	}

	// fmt.Printf("errors %#v\n\n", errors)
}
