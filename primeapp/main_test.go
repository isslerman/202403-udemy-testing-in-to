package main

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func Test_alpha_isPrime(t *testing.T) {
	// settings a table test with this fields.
	primeTests := []struct {
		name     string
		testNum  int
		expected bool
		msg      string
	}{
		// data to be tested
		{"prime", 7, true, "7 is a prime number!"},
		{"not a prime", 8, false, "8 is not a prime number because it is divisible by 2!"},
		{"negative number", -3, false, "Negative numbers are not prime, by definition!"},
		{"0 number", 0, false, "0 is not prime, by definition!"},
		{"1 number", 1, false, "1 is not prime, by definition!"},
	}

	for _, e := range primeTests {
		result, msg := isPrime(e.testNum)
		if e.expected && !result {
			t.Errorf("%s: expected true but got false", e.name)
		}
		if !e.expected && result {
			t.Errorf("%s: expected false but got true", e.name)
		}
		if e.msg != msg {
			t.Errorf("%s: expected %s but got %s", e.name, e.msg, msg)
		}
	}
}

func Test_alpha_Prompt(t *testing.T) {
	// save a copy of os.Stdout
	oldOut := os.Stdout

	// create a read and write pipe
	r, w, _ := os.Pipe()

	// set os.Stdout to our write pipe
	os.Stdout = w

	prompt()

	//close our writer
	_ = w.Close()

	// reset os.Stdout to what it was before
	os.Stdout = oldOut

	// read the output of our prompt() func our read pipe
	out, _ := io.ReadAll(r)

	// perform our test
	if string(out) != "-> " {
		t.Errorf("incorrect prompt: expected -> but got %s", string(out))
	}
}

func Test_alpha_intro(t *testing.T) {
	// save a copy of os.Stdout
	oldOut := os.Stdout

	// create a read and write pipe
	r, w, _ := os.Pipe()

	// set os.Stdout to our write pipe
	os.Stdout = w

	intro()

	//close our writer
	_ = w.Close()

	// reset os.Stdout to what it was before
	os.Stdout = oldOut

	// read the output of our prompt() func our read pipe
	out, _ := io.ReadAll(r)

	// perform our test
	// we will if it contains just a part of the string/code, not all.
	if !strings.Contains(string(out), "Enter a whole number") {
		//	if !strings.Contains(string(out), "+\nEnte") {
		t.Errorf("intro text not correct; got %s", string(out))
	}
}

func Test_alpha_checkNumbers(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty", "", "Please enter a whole number!"},
		{"exit", "q", ""},
		{"exit", "Q", ""},
		{"zero", "0", "0 is not prime, by definition!"},
		{"one", "1", "1 is not prime, by definition!"},
		{"seven", "7", "7 is a prime number!"},
		{"non prime number", "8", "8 is not a prime number because it is divisible by 2!"},
		{"negative", "-5", "Negative numbers are not prime, by definition!"},
		{"typed", "two", "Please enter a whole number!"},
		{"decimal", "1.1", "Please enter a whole number!"},
		{"greek", "ψφΦφΦ", "Please enter a whole number!"},
	}

	for _, e := range tests {
		// simulating what a user has typed
		input := strings.NewReader(e.input)
		reader := bufio.NewScanner(input)
		res, _ := checkNumbers(reader)

		if !strings.EqualFold(res, e.expected) {
			t.Errorf("%s: expected %s, but got %s", e.name, e.expected, res)
		}
	}
}

func Test_alpha_readUserInput(t *testing.T) {
	// to test this function, we need a channel, and an instance of an io.Reader
	doneChan := make(chan bool)

	// create a reference to a bytes.Buffer
	var stdin bytes.Buffer

	// it is like the user digit 1 enter q enter
	stdin.Write([]byte("1\nq\n"))

	go readUserInput(&stdin, doneChan)
	<-doneChan
	close(doneChan)
}

// func Test_isPrime(t *testing.T) {
// 	result, msg := isPrime(0)
// 	if result {
// 		t.Errorf("with %d as test parameter, got true, but expected false", 0)
// 	}

// 	if msg != "0 is not prime, by definition!" {
// 		t.Error("wrong message returned:", msg)
// 	}

// 	result, msg = isPrime(7)
// 	if !result {
// 		t.Errorf("with %d as test parameter, got false, but expected true", 7)
// 	}

// 	if msg != "7 is a prime number!" {
// 		t.Error("wrong message returned:", msg)
// 	}
// }
