package main

import (
	"fmt"
)

func main() {
	fmt.Println("// Samples from 'Miscellaneous Symbols and Pictographs': U+1F300 - U+1F5FF")
	printDetails("🍑")
	printDetails("🍕")
	printDetails("🍣")

	fmt.Println("")

	fmt.Println("// Samples from 'Supplemental Symbols and Pictographs': U+1F900 - U+1F9FF")
	printDetails("🥒")
	printDetails("🥑")
	printDetails("🦀")
	printDetails("🦁")
	printDetails("🦒")
	printDetails("🦐")

	fmt.Println("")

	fmt.Println("// Samples from 'Symbols and Pictographs Extended-A': U+1FA70 - U+1FAFF")
	printDetails("🩰")
	printDetails("🫐")
	printDetails("🫑")
	printDetails("🫒")
	printDetails("🫓")

	fmt.Println("")

	fmt.Println("// Samples from 'Miscellaneous Symbols': U+2600 - U+26FF")
	printDetails("⛺")
	printDetails("⛽")
	printDetails("⛳")
}

func printDetails(char string) {
	fmt.Printf("len(%s) = %d", char, len(char))

	fmt.Printf(",  dec:")
	for i := 0; i < len(char); i++ {
		fmt.Printf(" %d", char[i])
	}

	fmt.Printf(",  hex:")
	for i := 0; i < len(char); i++ {
		fmt.Printf(" %x", char[i])
	}

	// Convert to rune and print codepoint
	rune := []rune(char)
	fmt.Printf(",  codepoint: %U", rune)

	fmt.Println("")
}
