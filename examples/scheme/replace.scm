; scheme --load replace.scm < /dev/null
; Author: Matt Weaver
; Assignment: LA1-1
; Class: CS 354
; Due September 23, 2014
; File contents: replace function

; General comments: I was very comment-happy in this assignment. Since this is 
; 	the first time I have ever used Scheme, the comments helped greatly in 
; 	figuring out what each part of the program does, even when seemingly obvious

; replace - Replace every instance of an item in a list/tree with another item
; usage: (replace seq search-for replace-with)
; parameters: 
; 	seq          - The list/tree (sequence) to search
; 	search-for   - The thing to search for. Can be an atom or a list
; 	replace-with - What to replace each instance of search-for with.
;	               Can be an atom or a list (the type does not need to match
; 	               search-for)
; returns: A copy of seq, with every instance of search-for replaced with replace-with
(define (replace seq search-for replace-with) 
	(if (null? seq)     ; Verify seq exists and has contents, to deal with the 
						;   base case of seq being empty
		seq             ; Base case - seq is empty or null; return seq
		(if (list? seq) ; Check if seq is a list, so we don't accidentally
			            ; try to (car) a list element (another base case)
			(if (equal? (car seq) search-for)
				; Did we find a match? Cool. Recursively (replace) on just the tail
				(cons replace-with (replace (cdr seq) search-for replace-with))
				; Otherwise, Recursively (replace) on both the head (to deal
				; with trees) and tail
				(cons (replace (car seq) search-for replace-with) 
					  (replace (cdr seq) search-for replace-with)))
			; If this isn't a list, just check for a match and replace, accordingly
			(if (equal? seq search-for) replace-with seq))))