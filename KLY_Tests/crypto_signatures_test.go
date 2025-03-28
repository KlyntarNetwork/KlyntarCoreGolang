package tests

import (
	"fmt"
	"testing"

	bls "github.com/KlyntarNetwork/Web1337Golang/crypto_primitives/bls"
	ed25519 "github.com/KlyntarNetwork/Web1337Golang/crypto_primitives/ed25519"
	pqc "github.com/KlyntarNetwork/Web1337Golang/crypto_primitives/pqc"
	tbls "github.com/KlyntarNetwork/Web1337Golang/crypto_primitives/tbls"
)

func TestEd25519(t *testing.T) {

	myPubKey, myPrivateKey := ed25519.GenerateKeyPair()

	fmt.Println("PubKey is ", myPubKey)

	fmt.Println("PrivateKey is ", myPrivateKey)

	signa := ed25519.GenerateSignature(myPrivateKey, "Hello KLY")

	fmt.Println("Signa is ", signa)

	var isOk bool

	iteration := 0

	for i := 0; i < 20_000; i++ {

		isOk = ed25519.VerifySignature("Hello KLY", myPubKey, signa)

		iteration++

	}

	fmt.Println("Is ok =>", isOk)
	fmt.Println("Iteration =>", iteration)

}

func TestBliss(t *testing.T) {

	myPubKey, myPrivateKey := pqc.GenerateBlissKeypair()

	fmt.Println("PubKey is ", myPubKey)

	fmt.Println("PrivateKey is ", myPrivateKey)

	signa := pqc.GenerateBlissSignature(myPrivateKey, "Hello KLY")

	fmt.Println("Signa is ", signa)

	isOk := pqc.VerifyBlissSignature("Hello KLY", myPubKey, signa)

	fmt.Println("Is ok =>", isOk)

}

func TestDilithium(t *testing.T) {

	myPubKey, myPrivateKey := pqc.GenerateDilithiumKeypair()

	fmt.Println("PubKey is ", myPubKey)

	fmt.Println("PrivateKey is ", myPrivateKey)

	signa := pqc.GenerateDilithiumSignature(myPrivateKey, "Hello KLY")

	fmt.Println("Signa is ", signa)

	isOk := pqc.VerifyDilithiumSignature("Hello KLY", myPubKey, signa)

	fmt.Println("Is ok =>", isOk)

}

func TestBLS(t *testing.T) {

	// Generate keypair

	privateKey, publicKey := bls.GenerateKeypair()

	fmt.Println("Privatekey is => ", privateKey)

	fmt.Println("Publickey is => ", publicKey)

	// Generate signature

	message := "Hello KLY"

	signa := bls.GenerateSignature(privateKey, message)

	fmt.Println("Signa is => ", signa)

	// Now verify (True Positive)
	fmt.Println("Is ok with norm message => ", bls.VerifySignature(publicKey, message, signa))

	// Now verify with wrong msg (True Negative)
	fmt.Println("Is ok with norm message => ", bls.VerifySignature(publicKey, "Hello badass", signa))

	// Now generate more keypairs to test aggregation

	privateKey1, publicKey1 := bls.GenerateKeypair()
	_, publicKey2 := bls.GenerateKeypair()
	_, publicKey3 := bls.GenerateKeypair()

	signa1 := bls.GenerateSignature(privateKey1, message)
	// signa2 := crypto_primitives.GenerateBlsSignature(privateKey2, message)
	// signa3 := crypto_primitives.GenerateBlsSignature(privateKey3, message)

	aggregatedSigna := bls.AggregateSignatures([]string{signa, signa1})

	fmt.Println("Aggregated signa is => ", aggregatedSigna)

	// Aggregate pubkeys

	rootPubKey := bls.AggregatePubKeys([]string{publicKey, publicKey1, publicKey2, publicKey3})

	fmt.Println("RootPubKey is => ", rootPubKey)

	// Verify with threshold

	aggregatedPubOfSigners := bls.AggregatePubKeys([]string{publicKey, publicKey1})

	fmt.Println("Aggregated 0 and 1 is => ", aggregatedPubOfSigners)

	aggregatedPub23 := bls.AggregatePubKeys([]string{publicKey2, publicKey3})

	fmt.Println("Aggregated 2 and 3 is => ", aggregatedPub23)

	fmt.Println("Their sum => ", bls.AggregatePubKeys([]string{aggregatedPubOfSigners, aggregatedPub23}))

	fmt.Println("Is threshold reached => ", bls.VerifyThresholdSignature(aggregatedPubOfSigners, aggregatedSigna, rootPubKey, message, []string{publicKey2, publicKey3}, 2))

}

func TestTBLS(t *testing.T) {

	/*

		T = 2
		N = 3

	*/

	randomIDs := tbls.GenerateRandomIds(3)

	fmt.Println("IDs are => ", randomIDs)

	// Each group member do it individually

	vvec1, secretShares1 := tbls.GenerateTbls(2, randomIDs)
	vvec2, secretShares2 := tbls.GenerateTbls(2, randomIDs)
	vvec3, secretShares3 := tbls.GenerateTbls(2, randomIDs)

	fmt.Println("Vvec 1 ", vvec1)

	// Now derive rootPubKey

	rootPubKey := tbls.DeriveRootPubKey(vvec1, vvec2, vvec3)

	fmt.Println("RootPubKey is => ", rootPubKey)

	// Now imagine that members 1 and 2 aggree to sign something while member 3 - disagree. Generate partial signatures 1 and 2

	msg := "Hello World"

	secretSharesFor1 := []string{secretShares1[0], secretShares2[0], secretShares3[0]}
	secretSharesFor2 := []string{secretShares1[1], secretShares2[1], secretShares3[1]}

	partialSignature1 := tbls.GeneratePartialSignature(randomIDs[0], msg, secretSharesFor1)
	partialSignature2 := tbls.GeneratePartialSignature(randomIDs[1], msg, secretSharesFor2)

	fmt.Println("Partial signature 1 is => ", partialSignature1)
	fmt.Println("Partial signature 2 is => ", partialSignature2)

	// Aggregate them

	rootSignature := tbls.BuildRootSignature([]string{partialSignature1, partialSignature2}, []string{randomIDs[0], randomIDs[1]})

	fmt.Println("Root signature is => ", rootSignature)

	fmt.Println("Is root signature ok ? => ", tbls.VerifyRootSignature(rootPubKey, rootSignature, msg))

}
