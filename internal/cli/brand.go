package cli

import (
	"fmt"
	"io"
)

const batLogo = `       /\                 /\
      /  \__         __/  \
      \     \_______/     /
       \____/ ALFRED \____/`

func printBrand(w io.Writer, msg messages) {
	fmt.Fprintln(w, batLogo)
	fmt.Fprintln(w, msg.text("menu_title"))
}
