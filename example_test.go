package json_test

import (
	"fmt"
	"io"
	"log"

	"github.com/xsandr/json"
)

func ExampleScanner_Next() {
	input := `{"a": 1,"b": 123.456, "c": [null]}`
	sc := json.NewScanner([]byte(input))
	for {
		tok := sc.Next()
		if len(tok) < 1 {
			break
		}
		fmt.Printf("%s\n", tok)
	}
	// Fixme: think about having Error method
	//if err := sc.Error(); err != nil && err != io.EOF {
	//	log.Fatal(err)
	//}

	// Output:
	// {
	// "a"
	// :
	// 1
	// ,
	// "b"
	// :
	// 123.456
	// ,
	// "c"
	// :
	// [
	// null
	// ]
	// }
}

func ExampleDecoder_Token() {
	input := `{"a": 1,"b": 123.456, "c": [null]}`
	dec := json.NewDecoder([]byte(input))
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%v\n", tok)
	}

	// Output:
	// {
	// a
	// 1
	// b
	// 123.456
	// c
	// [
	// <nil>
	// ]
	// }
}

func ExampleDecoder_NextToken() {
	input := `{"a": 1,"b": 123.456, "c": [null]}`
	dec := json.NewDecoder([]byte(input))
	for {
		tok, err := dec.NextToken()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s\n", tok)
	}

	// Output:
	// {
	// "a"
	// 1
	// "b"
	// 123.456
	// "c"
	// [
	// null
	// ]
	// }
}
func ExampleDecoder_Decode() {
	input := `{"a": 1,"b": 123.456, "c": [null]}`
	dec := json.NewDecoder([]byte(input))
	var i interface{}
	err := dec.Decode(&i)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%v\n", i)

	// Output: map[a:1 b:123.456 c:[<nil>]]
}
