package parallel

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
)

type output struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func ExampleFunction() {
	// Run analyze up to 5 times in parallel
	f := Run(analyze, 5)

	// Accumulate the results
	var results []output
	f.OnResult(func(o interface{}) {
		results = append(results, o.(output))
	})

	// Report any error
	f.OnError(func(err error) {
		log.Fatal(err)
	})

	// Call the function 10 times
	for _, name := range []string{
		"James", "Mary", "Robert", "Patricia", "John", "Jennifer",
		"Michael", "Linda", "William", "Raphael"} {
		f.Call(name)
	}

	// Wait for all executions to complete
	f.Wait()

	// Print the results sorted by age
	sort.Slice(results, func(i, j int) bool {
		if results[i].Age == results[j].Age {
			return results[i].Name < results[j].Name
		}
		return results[i].Age < results[j].Age
	})
	for _, res := range results {
		fmt.Printf("%s: %d\n", res.Name, res.Age)
	}

	// Output:
	// Patricia: 21
	// Jennifer: 24
	// Raphael: 43
	// John: 57
	// William: 57
	// James: 58
	// Linda: 59
	// Robert: 60
	// Mary: 63
	// Michael: 69
}

func analyze(name interface{}) (interface{}, error) {
	resp, err := http.Get("https://api.agify.io/?name=" + name.(string))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s: %s", name, resp.Status)
	}
	var out output
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	out.Name = name.(string)
	return out, nil
}
