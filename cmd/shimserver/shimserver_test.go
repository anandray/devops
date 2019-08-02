package main

import "testing"

// test case is:
// 1. create account "accountxyz" with http://d0.wolk.com as shimUrl
// 2. user requests wolk://accountxyz/malala.mp4
// 3. wolk node does request of http://d0.wolk.com/malala.mp4
// 4. shimserver responds with content, signed
// 5. assuming signature checking, wolk node returns content to user, with signed tx in header, and includes tx in blockchain
// 6. user checks for tx in header
// THEN:
// 2b. a second request by user for wolk://accountxyz/malala.mp4
// 2c. wolk retrieves content from bucket
func TestShim(t *testing.T) {
	// 1. create account "accountxyz" with http://d0.wolk.com as shimUrl
	// 2. user requests wolk://accountxyz/malala.mp4 once
	// 6. user check for tx in header, check contentHash

	// THEN:
	// 2b. user requests wolk://accountxyz/malala.mp4 again

	// 2c. user receives NoSQL proof
}
