#!/Users/mhweaver/Documents/workspace/shebang/go-version/shebang -r "gcc -o !- !@ && !-; rm !-"


#include <stdio.h>

int main(int argc, char *argv[]) {
	printf("Hello, world!\n");
}