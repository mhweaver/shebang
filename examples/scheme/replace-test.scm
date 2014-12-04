#!/usr/bin/env shebang -d "!(scheme)"
;#! /Users/mhweaver/Documents/workspace/shebang/go-version/shebang -d "scheme --load !@ < /dev/null"
; scheme --load replace-test.scm < /dev/null
; Author: Matt Weaver
; Assignment: LA1-1
; Class: CS 354
; Due September 23, 2014
; File contents: test suite for the replace function (LA1.scm)

; General comments: For my tests, I used a few function calls with side effects.
; 	My main replace function does not use any, so I feel like I am following the
; 	spirit of the assignment. I just got tired of repeating the same display/
; 	pretty-print calls over and over, so I made a new function, to make life
; 	easier and prettier.

(load "replace.scm")

; Flag for if any tests failed
(define pass true)

; Display an expression/function, it's value, and what was expected.
; This pretty much just saves me a lot of typing and gives me a little
; extra scheme experience. Also, the tests just look nice.
;
; usage: (test expr expected-value)
; parameters:
; 	expr           - the expression to be evaluated
; 	expected-value - the expected value of expr
; returns: The value of expr
(define (test expr expected)
	; Evaluate and store the value of the expression in a local variable,
	; so it doesn't need to be eval'd multiple times
	(let ((output (eval expr user-initial-environment)))
		(newline)
		(pretty-print expr) ; Print the expression
						    ; (pretty-print), because '(1 2 3) is prettier
						    ; than (quote (1 2 3))
						    ; Not (pp), because that adds an unwanted \n
		(pretty-print ':)
		(newline)
		(pretty-print output) ; Print the function call value
		(newline)
		(pretty-print expected) ; Print what was expected
		(display " expected")
		(newline)
		; Set the pass flag, so we know if something failed or not
		(set! pass (and pass (equal? output expected)))
		output))

; Basic test for simple list
(test 
	'(replace '(1 2 3) 1 3) 
	'(3 2 3))

; Test for a simple tree, replacing a list in the list
(test 
	'(replace '(a (1 2 3) b (1 2) c (1 2 3) d (1 2 3)) '(1 2 3) '_)
	'(a _ b (1 2) c _ d _))

; Provided example test
(test
	'(replace 1 1 2)
	2)

; Provided example test
(test 
	'(replace '(a (b c) d) '(b c) '(x y))
	'(a (x y) d))

; Provided example test
(test 
	'(replace '(a (b c) (d (b c))) '(b c) '(x y))
	'(a (x y) (d (x y))))

; Test for tree with search-for pattern several levels down
(test 
	'(replace '(((1 2) (2 3)) 4 5 6) '(2 3) '(x y))
	'(((1 2) (x y)) 4 5 6))

; Replacement that shouldn't do anything, to verify the search-for pattern.
; as a single-element list isn't being removed from the list or something
(test 
	'(replace '(1 2 3) '(1) '(a))
	'(1 2 3))

; Basic test, to mix data types
(test 
	'(replace '(1 2 3) 1 'a)
	'(a 2 3))

; Replace with empty list
(test
	'(replace () 1 2)
	())

; Replace with empty list search-for pattern
(test
	'(replace '(1 2 3) () 1)
	'(1 2 3))

; Replace with empty list replace-with
(test
	'(replace '(1 2 3) 1 ())
	'(() 2 3))

; Replace with empty list search-for and replace-with
(test
	'(replace '(1 2 3) () ())
	'(1 2 3))

; Does something like (cdr '(1)) match ()?
(test
	'(replace `(1 2 ,(cdr '(1))) () 3)
	'(1 2 3))


(newline)
(display 
	(if pass
		"All tests passed"
		"A test failed!"))
