/*

mybot - Illustrative Slack bot in Go

Copyright (c) 2015 RapidLoop

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

type Fo struct {
  USD  float64
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "usage: mybot slack-bot-token\n")
		os.Exit(1)
	}

	// start a websocket-based Real Time API session
	ws, id := slackConnect(os.Args[1])
	fmt.Println("mybot ready, ^C exits")
	fmt.Sprintf(id)

	for {
		// read each incoming message
		m, err := getMessage(ws)
		if err != nil {
			log.Fatal(err)
		}

		// see if we're mentioned
		if m.Type == "message" {
			//postMessage(ws,m)
			// if so try to parse if
			parts := strings.Fields(m.Text)
			if len(parts) == 2 && strings.ToLower(parts[0]) == "s" {
				go func(m Message) {
					m.Text = getQuote(parts[1])
					postMessage(ws, m)
				}(m)
			} else if len(parts) == 2 && strings.ToLower(parts[0]) == "c" {
				cryptArray := strings.Split(parts[1], ",")
				for i := 0; i < len(cryptArray) ; i++ {
					go func(m Message) {
						m.Text = getCrypto(cryptArray[i])
						postMessage(ws, m)
					}(m)
				}
			}
		}
	}
}

func getQuote(sym string) string {
	sym = strings.ToUpper(sym)

	emo := fmt.Sprintf("")
	url := fmt.Sprintf("http://download.finance.yahoo.com/d/quotes.csv?s=%s&f=nsl1p2c1op&e=.csv", sym)
	concat := fmt.Sprintf(" ")
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}
	rows, err := csv.NewReader(resp.Body).ReadAll()
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}
	for i := 0; i < len(rows) ; i++ {
		if len(rows) >= 1 && len(rows[i]) == 7 {

			test := rows[i][3]
			if test[0:1] == "+" {
				emo = ":grin:"
			} else {
				emo = ":pensive:"
			}

		concat += fmt.Sprintf("*%s (%s) is trading at $%s, Change: %vperc(%s$)* %v\n", rows[i][0], rows[i][1], rows[i][2], test[0:4], rows[i][4], emo)
		}
	}

	return fmt.Sprintf(concat)
}

func getCrypto(f string) string {
	f = strings.ToUpper(f)
	client := &http.Client{}
	req, err := http.NewRequest( "GET", fmt.Sprintf("https://min-api.cryptocompare.com/data/price?fsym=%s&tsyms=USD", f), nil)
	if err != nil {
	  log.Fatalln(err)
	}
	req.Header.Add("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
	  return fmt.Sprintf(":wutface:")
	}
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	v := Fo{}
	err = decoder.Decode(&v)
	if err != nil {
	  return fmt.Sprintf(":wutface:")
	}

	return fmt.Sprintf("*%s is trading at $%.5f*", f, v.USD)
}
