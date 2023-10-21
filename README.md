# SFTPUserVerifier
Simple automated SFTP testing tool for single to multi-user scenarios

Feel free to copy, edit, fork, etc

Purpose:
This tool was created to automate the testing of new user's connection to an SFTP server, 
including the creation of new files and folders, chroot restrictions, and file removal capabilities.

Authentication:
While the current configuration supports username/password authentication, 
implementing public/private key authentication is straightforward.

QUICK START:
1. Download Go from https://go.dev/dl/
2. Clone this repository
3. Navigate to the repository folder
4. Update the configuration at the beginning of the code
5. Execute the command 'go run test.go'


Here's a rundown what you need to edit:
1. Server IP
2. Server Port
3. Users to test
    users    = []UserInfo{
		{User: "user1", Password: "password1"}, 
		{User: "user2", Password: "password2"},
		// ... add more users as needed
	}
4. ServerTestFilePath - Where the test file is uploaded - Should be writable by user like '/files/test.bin'  
5. SSH Configuration: Update the SSH configuration parameters. Check naming from https://github.com/sshnet/SSH.NET
            KeyExchanges: []string{"ENTER KEX HERE"}, // example "diffie-hellman-group16-sha512"
			Ciphers:      []string{"ENTER CIPHER HERE"},
			MACs:         []string{"ENTER MAC HERE"}