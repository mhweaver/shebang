shebang:	shebang.go
	go build -o $@

install:	shebang
	cp $< ~/bin/

remove:
	rm ~/bin/shebang

usage:	shebang
	./$<

clean:
	rm shebang