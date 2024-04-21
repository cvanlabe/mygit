package main

import (
	"bytes"
	"compress/zlib"
	"flag"
	"fmt"
	"io"
	"os"
)

// Implements the git init command
//
// It creates the .git/ folder and its children:
//
//	.git/HEAD
//	.git/objects
//	.git/refs
func initCmd() {
	foldersToCreate := []string{".git/", ".git/objects", ".git/refs"}

	for _, folder := range foldersToCreate {
		err := os.Mkdir(folder, 0750)
		if err != nil && !os.IsExist(err) {
			error := fmt.Sprintf("Failed to create folder '%s': %s", folder, err)
			fmt.Fprintln(os.Stderr, error)
			os.Exit(1)
		}
	}

	filesToCreate := ".git/HEAD"
	headContent := []byte("ref: refs/heads/main\n")

	err := os.WriteFile(filesToCreate, headContent, 0644)
	if err != nil {
		error := fmt.Sprintf("Failed to create HEAD: %s", err)
		fmt.Fprintln(os.Stderr, error)
		os.Exit(1)
	}
	fmt.Println("Initialized .git directory")
}

// git cat-file -p <blob_sha> pretty-prints the contents of a git object in the .git/objects/ folder
// Each blob is stored on disk with:
// - a filename based on its 40-character SHA-1 hash, where:
//   - the first 2 characters are a subdirectory in the .git/objects/ folder
//   - the remaining 38 characters is its filename
//
// - zlib compressed contents
//
// The blob object has a header before the real content:
//
//	blog <size>\0<actual content>
//
// It's therefore important that if we pretty-print, we discard that header first.
func catFile(args []string) {
	flag := flag.NewFlagSet("git cat-file", flag.ExitOnError)
	var (
		pprint = flag.Bool("p", false, "pretty-print the contents of <object> based on its type")
	)
	flag.Parse(args)
	args = flag.Args()

	if *pprint {
		//fmt.Println("pretty-print enabled")
	}

	if len(args) <= 0 {
		fmt.Fprintln(os.Stderr, "usage: git cat-file [-p] <blob_sha>")
		os.Exit(1)
	}

	// eg: object 0a5159e4fd9efdc3530c880fa15b672f08d47421
	// would be stored in .git/0a/5159e4fd9efdc3530c880fa15b672f08d47421
	object := args[0]
	dir := object[:2]
	filename := object[2:]
	file := fmt.Sprintf(".git/objects/%s/%s", dir, filename)

	fileContents, err := os.ReadFile(file)

	if err != nil {
		error := fmt.Sprintf("Failed to read '%s': %s", file, err)
		fmt.Fprintln(os.Stderr, error)
		os.Exit(1)
	}

	bytesReader := bytes.NewReader(fileContents)
	zReader, err := zlib.NewReader(bytesReader)

	if err != nil {
		error := fmt.Sprintf("Failed to decompress content of '%s': %s", file, err)
		fmt.Fprintln(os.Stderr, error)
		os.Exit(1)
	}

	defer zReader.Close()

	decompressedContents, _ := io.ReadAll(zReader)
	headerEndOffset := findNullByteIndex(decompressedContents)

	fmt.Print(string(decompressedContents[headerEndOffset+1:]))

}

// hashObject -w <file> reads a provided file
// computes the SHA-1 hash of its content,
// writes the header+actual content to the file in the .git/objects folder:
//
// The content will be:
//
//	blob <size in bytes>\0<actual content>
func hashObject(args []string) {
	flag := flag.NewFlagSet("git hash-object", flag.ExitOnError)
	var (
		write = flag.Bool("w", false, "Actually write the object into the object database")
	)
	flag.Parse(args)
	args = flag.Args()

	file := args[0]

	fileContent, err := os.ReadFile(file)
	if err != nil {
		error := fmt.Sprintf("Failed to read file '%s'. Error: %s", file, err)
		fmt.Fprintln(os.Stderr, error)
		os.Exit(1)
	}

	byteSize := len(fileContent)

	// Write the header: `blob <byteSize>\0<actual file content>`
	blobContents := []byte(fmt.Sprintf("blob %d\x00%s", byteSize, fileContent))

	// 40 character SHA-1 hash is based on the entire uncompressed content WITH header
	hash := sha1Hash(blobContents)

	fmt.Println(string(hash))

	if *write {
		// filename for objects database is based on the hash
		objectFolder := hash[:2]
		objectFile := hash[2:]

		// create the object directory
		err := os.Mkdir(fmt.Sprintf(".git/objects/%s", objectFolder), 0750)
		if err != nil && !os.IsExist(err) {
			error := fmt.Sprintf("Failed to create folder '%s'. Error: %s", objectFolder, err)
			fmt.Fprintln(os.Stderr, error)
			os.Exit(1)
		}

		// create the file to write zlib compressed data to
		compressedFile := fmt.Sprintf(".git/objects/%s/%s", objectFolder, objectFile)
		compressedBlobContents, err := os.Create(compressedFile)

		if err != nil {
			error := fmt.Sprintf("Failed to create compressed file '%s'. Error: %s", objectFile, err)
			fmt.Fprintln(os.Stderr, error)
			os.Exit(1)
		}

		compressedWriter := zlib.NewWriter(compressedBlobContents)
		defer compressedWriter.Close()

		_, err = compressedWriter.Write(blobContents)
		if err != nil {
			error := fmt.Sprintf("Failed to write zlib compressed data to '%s': %s", compressedFile, err)
			fmt.Fprintln(os.Stderr, error)
			os.Exit(1)
		}

	}
}

// Usage: your_git.sh <command> <arg1> <arg2> ...
func main() {

	flag.Parse()
	arguments := flag.Args()

	if len(arguments) == 0 {
		fmt.Fprintln(os.Stderr, "usage: git <command> [<args..]")
		os.Exit(1)
	}

	switch command, commandArgs := arguments[0], arguments[1:]; command {
	case "init":
		initCmd()

	case "cat-file":
		catFile(commandArgs)

	case "hash-object":
		hashObject(commandArgs)

	default:
		fmt.Fprintln(os.Stderr, "Not yet implemented git command")
		os.Exit(1)
	}
}
