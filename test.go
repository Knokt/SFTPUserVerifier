package main

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"os"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// Define userInfo struct
type UserInfo struct {
	User     string
	Password string
}

var (
	// Edit these
	server    = "IP ADDRESS HERE"
	port      = "PORT NUMBER HERE"
	sshServer = fmt.Sprintf("%s:%s", server, port)
	users     = []UserInfo{
		{User: "user1", Password: "password1"},
		{User: "user2", Password: "password2"},
		// ... add more users as needed
	}
	serverTestFilePath    = "SERVER_FILE_PATH_HERE" // Example: "/files/test.bin"
	localTestFilePath     = "local_test.bin"        // Leave default or enter custom name
	serverToLocalFilePath = "SERVER2LOCAL.bin"      // Leave default or enter custom name
	testFileSize          = 1024 * 1024             // 1024kb = 1mb
)

func getSSHConfig(user UserInfo) *ssh.ClientConfig {
	return &ssh.ClientConfig{
		User: user.User,
		Auth: []ssh.AuthMethod{
			ssh.Password(user.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		// Edit these
		Config: ssh.Config{
			KeyExchanges: []string{"ENTER KEX HERE"},
			Ciphers:      []string{"ENTER CIPHER HERE"},
			MACs:         []string{"ENTER MAC HERE"},
		},
	}
}

func main() {

	for _, userInfo := range users {
		fmt.Println("******************************")
		fmt.Printf("%s - testing\n", userInfo.User)
		sshClientConfig := getSSHConfig(userInfo)

		client, err := ssh.Dial("tcp", sshServer, sshClientConfig)
		if err != nil {
			fmt.Println("Error dialing the Server: ", err)
			fmt.Println("TEST FAILED!")
			continue
		}
		defer client.Close()

		sftpClient, err := sftp.NewClient(client)
		if err != nil {
			fmt.Println("Error creating sftp instance: ", err)
			fmt.Println("TEST FAILED!")
			continue
		}
		defer sftpClient.Close()

		// Generate a 1mb (1024 kb) test file with random data
		err = generateRandomBinaryFile()
		if err != nil {
			fmt.Println("Error generating test binary file. Skipping rest of the loop execution", err)
			fmt.Println("TEST FAILED!")
			continue
		} else {
			fmt.Println("Test file generated successfully")
		}

		// Comment out tests that you don't want to run
		tests := []func(*sftp.Client) (bool, string, string){
			testSaveToChroot,
			uploadFile,
			downloadFile,
			compareFiles,
			testRemoveFile,
		}

		totalTests, completedTests := len(tests), 0

		fmt.Println("Proceeding to testing: ")
		for _, test := range tests {
			completed, testName, errMsg := test(sftpClient)
			printTestResult(completed, testName, errMsg)
			if completed {
				completedTests++
			}
		}

		fmt.Printf("Test finished %d / %d tests completed.\n", completedTests, totalTests)
	} // end of loop - users
} // end of main

func printTestResult(completed bool, testName, errMsg string) {
	status := "Failed"
	if completed {
		status = "Completed"
	}
	fmt.Printf("%s %s - %s\n", testName, status, errMsg)
}

func generateRandomBinaryFile() error {
	// Create a new file or overwrite the existing file
	file, err := os.Create(localTestFilePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Create a byte slice to hold the random data
	data := make([]byte, testFileSize)

	// Fill the byte slice with random data
	_, err = rand.Read(data)
	if err != nil {
		return fmt.Errorf("failed to generate random data: %w", err)
	}

	// Write the random data to the file
	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}

	return nil
}

func uploadFile(client *sftp.Client) (bool, string, string) {
	testName := "uploadFile"
	// Open the source file
	srcFile, err := os.Open(localTestFilePath)
	if err != nil {
		return false, testName, fmt.Sprintf("Failed to open source file: %s", err)
	}
	defer srcFile.Close()

	// Create the destination file on the SFTP server
	dstFile, err := client.Create(serverTestFilePath)
	if err != nil {
		return false, testName, fmt.Sprintf("Failed to create destination file: %s", err)
	}
	defer dstFile.Close()

	// Copy the contents of the source file to the destination file
	_, err = dstFile.ReadFrom(srcFile)
	if err != nil {
		return false, testName, fmt.Sprintf("Failed to copy file: %s", err)
	}

	return true, testName, "File uploaded successfully"
}

func downloadFile(client *sftp.Client) (bool, string, string) {
	testName := "downloadFile"
	// Open the remote file
	remoteFile, err := client.Open(serverTestFilePath)
	if err != nil {
		return false, testName, fmt.Sprintf("Failed to open remote file: %s", err)
	}
	defer remoteFile.Close()

	// Create a new local file
	localFile, err := os.Create(serverToLocalFilePath)
	if err != nil {
		return false, testName, fmt.Sprintf("Failed to create local file: %s", err)
	}
	defer localFile.Close()

	// Download the file by copying the contents of the remote file to the local file
	_, err = io.Copy(localFile, remoteFile)
	if err != nil {
		return false, testName, fmt.Sprintf("Failed to download file: %s", err)
	}

	return true, testName, "File downloaded successfully"
}

func compareFiles(sftpClient *sftp.Client) (bool, string, string) {
	testName := "compareFiles"
	// Open local_test.bin file
	localTestFile, err := os.Open(localTestFilePath)
	if err != nil {
		return false, testName, fmt.Sprintf("Failed to open local file: %s", err)
	}
	defer localTestFile.Close()

	// Open server2local_test.bin file
	server2localTestFile, err := os.Open(serverToLocalFilePath)
	if err != nil {
		return false, testName, fmt.Sprintf("Failed to open local file: %s", err)
	}
	defer server2localTestFile.Close()

	// Compute the SHA-256 hash of the local file
	localHash := sha256.New()
	if _, err := io.Copy(localHash, localTestFile); err != nil {
		return false, testName, fmt.Sprintf("Failed to read local file: %s", err)
	}

	// Compute the SHA-256 hash of the remote file
	server2localHash := sha256.New()
	if _, err := io.Copy(server2localHash, server2localTestFile); err != nil {
		return false, testName, fmt.Sprintf("Failed to read remote file: %s", err)
	}

	// Compare the hashes
	if fmt.Sprintf("%x", server2localHash.Sum(nil)) != fmt.Sprintf("%x", localHash.Sum(nil)) {
		return false, testName, "Files are not identical"
	}

	return true, testName, "Files are identical"
}

func testRemoveFile(client *sftp.Client) (bool, string, string) {
	testName := "testRemoveFile"
	// Attempting to remove the remote file
	err := client.Remove(serverTestFilePath)
	if err != nil {
		switch err := err.(type) {
		case *os.PathError:
			if os.IsNotExist(err) {
				return false, testName, "Failed to remove: file doesn't exist"
			}
			if os.IsPermission(err) {
				return false, testName, "Failed to remove: permission denied"
			}
		}
		return false, testName, fmt.Sprintf("Failed to remove file: %s", err)
	}

	return true, testName, "File removed"
}

/*
* This function works backways in terms of test completed
* it return true if the first step, that is creating a new file in chroot fails
* Other cases it returns false to indicate there's an issue with the chroot restrictions
 */
func testSaveToChroot(client *sftp.Client) (bool, string, string) {
	testName := "testSaveToChroot"
	// Attempting to create a new file in chroot directory
	file, err := client.Create("test.bin")
	if err != nil {
		return true, testName, fmt.Sprintf("result: %s. *This is expected outcome*", err)
	}
	defer file.Close()

	// Attempting to write some data to the file
	_, err = file.Write([]byte("test data"))
	if err != nil {
		return false, testName, fmt.Sprintf("result: Failed: %s", err)
	}

	// If the above operations succeed, it's unexpected in a chroot environment
	return false, testName, "*Warning* Saving to chroot enabled"
}
