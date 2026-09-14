# FindWord

A structured Golang application designed to parse text and identify different phrases among a given set of words.

## Features

* **Phrase Identification:** Processes raw text to isolate and identify specific phrases or word groupings.
* **Structured Architecture:** Built with a clean Go structure, making it easy to scale and maintain as a foundational backend utility.
* **Sample Data Integration:** Includes a dedicated data module to test text extraction immediately upon cloning.

## Project Structure

* **`main.go`**: The core application logic containing the key libraries and methods for text processing.
* **`data/`**: Directory containing sample text files used for testing the extraction functions.
* **`NOTES.md`**: Project requirements, development notes, and planning documentation.
* **`LICENSE`**: MIT License file.

## Getting Started

### Prerequisites

* [Go](https://go.dev/dl/) installed on your local machine.

### Installation

1. Clone the repository to your local environment:
```bash
git clone https://github.com/SithilSemitha/findword.git

```


2. Navigate into the project directory:
```bash
cd findword

```



### Execution

Run the application directly using the Go CLI. The application will automatically look for the sample data files mapped in the `data` directory.

```bash
go run main.go

```

## Usage

To test the extraction logic, add your target text files into the `data/` directory. Modify `main.go` if necessary to point to your specific file names. When executed, the application will output the identified phrases to the console based on the defined matching logic.

## Author

**Sithil De Silva**

AI Engineering Student | [GitHub Profile](https://github.com/SithilSemitha)

## License

This project is licensed under the [MIT License](https://github.com/SithilSemitha/findword/blob/main/LICENSE).
