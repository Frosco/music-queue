### **Project Brief: Go Music Queue**

  * **Project Title**: Go Music Queue
  * **Version**: 1.0
  * **Date**: 2024-10-27
  * **Author**: Mary, Business Analyst
  * **Stakeholder**: User

### 1\. Project Overview

The "Go Music Queue" is a simple, command-line application designed for personal use. Its purpose is to manage a queue of music albums that the user wants to listen to. The application will be written in the Go programming language and will prioritize simplicity in its design and operation. It will run locally, storing the album queue in a plain text file.

### 2\. Core Requirements

#### 2.1. Technology Stack

  * **Programming Language**: Go
  * **Data Storage**: A single `.txt` file (e.g., `queue.txt`) will be used to persist the album list. Each line in the file will represent a single album.

#### 2.2. Album Representation

  * An album will be represented as a single string in the format: `Artist Name - Album Title`.

#### 2.3. Duplicate Detection

  * The application must prevent duplicate albums from being added to the queue.
  * The check will be performed as a **case-insensitive**, exact match of the entire album string. For example, "Led Zeppelin - IV" and "led zeppelin - iv" should be considered identical.

### 3\. Command-Line Interface (CLI) Specification

The application will be controlled via a series of commands.

#### 3.1. `queue import <filename.txt>`

  * **Action**: Imports a list of albums from a specified text file. The file must have one album per line in the `Artist Name - Album Title` format.
  * **Logic**: Each line from the input file is read. Before adding to the main `queue.txt` file, a case-insensitive duplicate check is performed.
  * **User Feedback**: The tool will report the number of albums successfully imported and the number of duplicates found and skipped.
      * *Example Output*: `Import complete. Added: 15 albums. Skipped: 3 duplicates.`

#### 3.2. `queue add "Artist Name - Album Title"`

  * **Action**: Adds a single album to the queue directly from the command line. The album string must be enclosed in quotes.
  * **Logic**: Performs the same case-insensitive duplicate check before appending the new album to the `queue.txt` file.
  * **User Feedback**: Confirms whether the album was added or was a duplicate.
      * *Example Success*: `Added "Daft Punk - Discovery" to the queue.`
      * *Example Duplicate*: `Duplicate found. "Daft Punk - Discovery" is already in the queue.`

#### 3.3. `queue next`

  * **Action**: Provides the user with their next album to listen to.
  * **Logic**:
    1.  Read all albums from the `queue.txt` file.
    2.  Randomly select one album from the list.
    3.  Display the selected album to the user.
    4.  Remove the selected album from the list and overwrite the `queue.txt` file with the updated, shorter list.
  * **User Feedback**:
      * *On Success*: `Now Listening: King Gizzard & The Lizard Wizard - PetroDragonic Apocalypse`
      * *If Empty*: `The queue is empty!`

#### 3.4. `queue list`

  * **Action**: Displays the entire contents of the music queue without modifying it.
  * **Logic**: Reads and prints every line from the `queue.txt` file.
  * **User Feedback**: Displays a numbered list of all albums in the queue.
      * *Example Output*:
        ```
        Current Queue:
        1. The Mars Volta - De-Loused in the Comatorium
        2. Radiohead - In Rainbows
        3. ...
        ```

#### 3.5. `queue count`

  * **Action**: Reports the total number of albums currently in the queue.
  * **Logic**: Counts the number of lines in the `queue.txt` file.
  * **User Feedback**: A simple text message.
      * *Example Output*: `There are 7 albums in the queue.`
