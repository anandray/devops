// Copyright 2018 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package crypto

import (
	"bytes"
	"crypto/sha256"
	"io"
	"math/rand"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

var expectedTransformedHex = "e3f4432569de65c70016d4606a7ee5dc458be20bacafdd28b6b61c194c73f74ec84a2a8ed9dc623ca3b132cc8aceaba0b608641d186511cdb0b9449db61d39f3bc027ab67b80164793694a8668ba59147f831ae92eee87849ff03eea2c66f15f46ff60a127b0e0c3d3f6c924c265088ac15469d08671231c0b96f5fe0eb2feb4e31784c3b7646f1d38f05ad538108ff13b60f9b9a6dba3f477badac6a0e0deb757c5c91cce72d80efe25b0576c4184269822fc36d4d32ce51121868466553f3b01f6ac7fae899170e1916a81c4e5b5c33ad081214339a69a501270aa36f75d35156aafdc8a46cacef9def7e4424f35827b254349c2fcc1e5bfabab543b2b98b924f3a1ebd7ce361ea9a16a0eb34f02c192fff1538815a311ca8d686a2c60bdfa7cb15b9c05901bacf28b4b23d7e6a54d440047d37f58424595c5ed774ef2656525177ab4faca53f2889cd27172dc764db8d5f4057de05bcfd1b9c70e4811ee02951c5ea372e22ca498e6a2f218a6ee71ba38c3dfa94259923b17b180798c99322e704975702d066512c3efdabd0cca1a565a98ecaa56353631a1347d64e8a532c6ad8dc7b926652790fda0e20694ff642820ed3540fdc51c8a1f66863cfe661721eeab0ab82f2b42c343da4c82af83144a3510f0e63858149aa284a56355735829c46eb295cf91f90e91b0e21b45a1f99be89b6fb34f1945004f284cedcfead7e7a6ead573c540bacc47350598407288425bbbfa87b85b98c55725f7c80385617e3cfa103310cd27650e540f0ee6a6e8be84f94fa8c48b3f4123316ce3ac8be65db4406e77764866e4dbc217be5e62c859426066f2ab3c59bf6a99d239bed5279b53e8285c6db35de4faba93bfa0410289def7d50734e6ffacfa6bc01f78e7a3d47a6e84f071c2160c0f76b7fdb3f0985bbd2d5e5b0d252cf7cdde648935e08baf58656ba4ad7eea2b3223f1a049df71c1fd83b1b790683187936f21c291e9f4e103ab068865b6b077bcaae06514af572bdc536709a7c6e32f23a1875292353899087489543d4c9651f24557fa8c6ed9959aaf481d2c12935cc11b611669339763b479ed3265cbf8cec21a3ced7cfa390f838d76b87969a59400065f5d33fc6e4906f182985e2a174c473c4af49945f235d55f2e8dcd9a03402e34ed1963e6b816ed20f6a2ea2ccb65afcb2a6fb44b47c32c180ee91a4f8fa54e77c9ef627636b12bdb287d64745f0c1c80d2f93c8e4bc7bb25f2cfa44dc77e9710dcff79c3aaa2777d2e604640d4e041c261db1813faf42c08cf780cbd35a1d08e7e8ca32094a3f26b8927e1885bff6db935ad45ecb27df3a4304ca3e35da15ae4f7fd73462ff29c3a7986b6bafe5fcdb08ea3d4ab1d9a8438048dd62767f9a8acc26844e67a33cbd1b0b422b4c2171b0fa28bbf30f6ca28993945ffa482fea8c416c8144457239bd08d26168410b607ec74fccf65ee4fdac0132383b92284bfef1a456edf5fab9929616c8e7ff327eacff0dcc3119cfeea19743e66302bee13ab900740f1e2e5273b47b062b4f6586b0abaffd9c9b91cc2c575a0980e6d46762ef3ae83971058690651d95e133a14e599b4c42df1c6d2476c77b7ed01de6fd61ee31d6e3f2910dea29a7eb44c82f6049d8953e076bb18e420ed59cd546abce5bc029d835fac800be7c50435c775fdd73f1ce038571c52dc77f62d9062312c1148ace1770dd90751c6d635b4ed9e398d16d34d47bc0152b309b618381e5645975991b025d904f73345d1975146de74ae3fb7b20c1a2e0c669252253a8dbf211830cbdc23b03b79be768dfbfbe072c97b49586cf5e7bac632e84e8e21909e4987aa45a6139fe146aba04937446df20dcca4d13f18da0e81761b7c38cb07fee64af40cfe00e6478376700f3ccdb07acbc9c3efc8924f8c343957fff26c033a51ea30de003af11b08f9e77ab0c3010ce553894d77c02a6344bdce4fe8260b41e4caf1086e66bec14bba35657c16740b77eda65d8ec3e599d64065639598a10c87a475a0b3822427e89151bebc979a9c7fb07dd6aa9167f5d5f0ecd2d6ddbe72b2a198e43f6f5df95c3dd84e4e240a712fa584948a762a0537c25c6fe3357ca89833aff7c416a65a8d837f7cd268b96a2f44248843d2baa066011021ff6fbee2d7bc923983ba006d3df9b1b997b36a2d1e460bae8cbbe7e59ecbc2cc5da1cff8a1d27452c0fc56d2dbc9ae743fe73dadcf960eb91d0275324471cc189ec3d29419948b95ede79279b8ee6ab8836fe444c6c1b2bc4fc2f949b2512101de21ea9408ffd06ee499de3d2004441356c835841484ec39e0547d6b96bdf05dc537d9d4f182d04ee749504486817a18619cf89022061c43ff4d069921eb18874339178c888d3f9145b6acd1cd97b8bf779b99e1e72011db1fb2930aecccdf98e6eba15161fcade889629cd4713f5b2366083431c229b0c544d36e21fe171ac827ec20f8ba947537c489fbdc7ad4cc2afc5b17c4ecea1fc45f7a432a89a332595801392b6268dd08efeb89dfa80567da7722bfcde3e3761d4fe82acaa04ac8b59b40821fdabff2b29cc6bbfe2fcd8f2a1cb5a024d3e39f2f81e5cec92284fb2214b107e50fd9b95117141ce03823cfbee9d46116f4f278b6d4e8883fbde9ea20e93ea5362b7d310153e8e0fb4530476a3c475d6a42cb7881889d57e8e5d423d7794e80983029615a0d0334834be86bda1def04a1254a79578f05a2aac7bae9f8c26cc769ffacdd8f8785713c4e267823b4811344efe5134136d71cafec30a0c2188d89fb6ce62250771ac43bb240666c7e240be7d895e649055129c8f01c2cdf66379eb808480b6c443601bf7e4c0065bc9fe013977804a8ccdffafe8e068d1e89fd51b1c59436fab252df7c28ea94df4e49eb92c02ed812e3116ece9e2f98cf0c689d9b3b8f476c3677a64bda4c054b30ecf1e420c82647bffe9a725e6585ad95f407ea941ddbb18d7338c6eb8eb1388f7f298da34ec02e3a6d3700eb3a2706582743417ad17e679ea79d9d8f4fef23773948dbe4a30b7f03f90adb13cbd5ba2a8f9c4b2ca6eec930b694f81a7756eefbbcd837fcb56eeb5ae106f5635bc8e5bcfd493526565d131e66a173f57b2590ffec85fdb9fe2b5e34f1bc368571f57a5531d99bf1d92632d28cdc3cab72a2fa07a4465f22c96fb19a401851139bf1b9c7699a69e036f2d7de68c09110e0375256c27a83080427c8a74f1f712207936795fb8c00cc36e7496a55d6b03dd44f5531fe4c34bf644630b284597d41220530145856502d7a4f16404fa58edab5fc86d712c75ce912dc318e2d8f271e6233523560888410ef40cfc2d96e6555c583e4cf55c31cd6ad9cf555ed807b9b6eea8fc19c28778e33c7b6e742c263594a5fd5886083b064ff2ae10c7a317c10f896f065a3acd23c0a1d75790407362a3869bfee84fd6d93554ec6d2a2996a4a024d8dd3a6a8b466a16c66702b741cef70b0203b1e1c455bacfe3630424587291d521d152d317af1c781fcf06436b8bce00af406da7912f47575701b6e60f2e20a2659f943d5912f2c68434df61ded9830035318fbe7cd0aebcbc23090830aaefc85de5335f52c697fe03a5b2730a1c7804c2b4ca63d8fd0a0b2a6b98320ab8cd5089028c9f814a6b42b3dff98e69eeef5ac9e077969dcaab60fea8a2e41cbd2ab8c5b1c6538e4e40c8e40060c3c921abc518743b57ce8b7b8e71cb05935c2970ccc1df90a9f1b9cb10171e4958d5e45c63f5416958099003e2ee1750f786d597fe2a118489c578b9ee9c45d483d86653cc3df406fcbdfc24e6625cec50da80c496a7af572969b9a91fb7fe0834378c28bd209833994b4e1890cff4d25efbaefaad566326fe3234a173c0f9292e9cfb5e74b33b8dd23502560578fc1b39e568245dc080cd0d9b3d7c9dd26156f4fb6d4a0a5d2c802cae22c0cdf7357eb79f3c06b55a0a529aeea96fcf36223d060a1a6dee4dd01f86307632e4fdb9d45b259ea81f6217a7eb0163bac90f022adece07738d15162cd35c06971c1faa55f32da46caa365e5a31dd0cb0f5ddc6b78cd729f8c512e5493e975af7fd4d78b03fdeac469a4bcb831916655f54c6149cd111b9c17ece415cb5ed12d01e2a189a16c5d7b62b8e6591a70d9c300b70ffb092384654fa73608b6959d0aa32f60c70c63e30683dda3b949432506f531e7f0904abd54e7f6ee5e0ae10000cd5abfdf6130fd059ec45f8e3ec2605eb30923588b8b7960d669521d889cd2b228bd8ac3d5d3eedc0269213d644b46f720a9ae9ec9ecd2dc334c6039c09239a84e8c1fd6cb0b88f0246c4a76b0db462620ac9eaef4453e1ae0d461a63be862f44141131945fe2d4bea184ae20d7ea8bc3cce7f671d1041b483f5f480bfa0bda4f168ca22f4f464027e9cc8dc6b71740de7bac95e26f6c035fd71c818ef1e555ef93991d118d53a1255c7b2368b521f3aad3be95794cb8697a9d21000a8417a5985a54eb35fe8caeb5b5f8616f81d6cd6aeff419040437beeeb4297560020e1484a067a525e999f54aa9319fb2bbf9689e0938076f333738b56fb14cb34c28a9ed14fdafba3d9501dce650217f07b3f577a390bc52ffe10e093c5dff0574559937959dcf831818567ca5fd2b3dee6cb9dc7fdb24cb077a479fe97652de620a3ff7fb81cee79117fbfb62824436753fc6e8b52abaee9309a12ab6bb2b9db7c05eace15d194b84efc7958abffb9f6d631e84a64625c91e7f326e97658389f145c6403e6ff75f494cd3068fa85cd69ec3a128a288f0f19c2f88f9ebf1bc9b31d51292f34767eb2e4618522ec3009959becaa3473f7f085a1130072742fff479b016a0c2e509119d56ea89a234ef64f19aa63bfa3e68cb748267a8ba56310a667d4837145b998cc999dc191db6a5e150c2c76f11a81829f8a6519823b62db031453fd863802dd4b71f057706d69f6254298b7ad43291069874991508794c224ef831d26285d02a7d3597c433b857ac76f624e49d41462f57bcfb3f4fa6eb481f4510031f2a4127f696883907097eefd5d7ddcd2227ab5e77c7d3a8169f59d19859155c55e85e9498bc86e85aa6defcaff3a12aba4c9f4fb6a8407ae3152902b51aa189a079d7430047911d237e4d0ef16c07801c3e060f049519c77354e366c978157ad968e674d0d37e7fe399aea9bdf1dd8a923c1cc6f49a195cb81a9fc305c86ad49845f099af340e6c63c99bb1d6a845e1640b37dca12743b9d4b8c958eb63a225e6c26dbbbd8b46270fb79992e44b99b42d03344a4718d51b437082c4048c2564ee6972f54002b8053b1fec47711d634d13dfc3a45a3a96d1c5c3862d23697f19dcd0e7f853088d95e400ff85f1dbe5d27692498e78dbab44a40427b71c47ce28715e45255de6796aed46a93b510b9c23f5a125d52dd9befd187f67210031f6a2ffb6a6e2a57354b9b76ca27cfbe13c2396e40937ede1e0554745c62fac72fe3a9f4e43db8780a6a296e84ad22ace1969be11b01495eb51782acaec863d82d7b4d533f7432ed0447684dade0e9782dae858739c928ea4e1cc5d37c1758e9ed4a502fe1b1e7789e0dffcd6f1a64f47818945e0b6d7bae7edaa7ca1c01c55438668b8452b8ee181abeccf045ddc65ac195f0ea9b178d18cb42add6218d25166d5f0ab46c43217e93c847e814536bd3a63f26d0a4e99eb13c3a484670289bc33772ef80741ada4840f21f0ef61a542c637851b5e4bc2ab83d5d1d5a5b645ec64082fcf897a870f4e0a1570ad1a6c2d52065caca8820ac6a0b"

var hashFunc = sha256.New
var testKey Key

func init() {
	var err error
	testKey, err = hexutil.Decode("0x8abf1502f557f15026716030fb6384792583daf39608a3cd02ff2f47e9bc6e49")
	if err != nil {
		panic(err.Error())
	}
}

func TestEncryptDataLongerThanPadding(t *testing.T) {
	enc := New(testKey, 4095, uint32(0), hashFunc)

	data := make([]byte, 4096)

	expectedError := "Data length longer than padding, data length 4096 padding 4095"

	_, err := enc.Encrypt(data)
	if err == nil || err.Error() != expectedError {
		t.Fatalf("Expected error \"%v\" got \"%v\"", expectedError, err)
	}
}

func TestEncryptDataZeroPadding(t *testing.T) {
	enc := New(testKey, 0, uint32(0), hashFunc)

	data := make([]byte, 2048)

	encrypted, err := enc.Encrypt(data)
	if err != nil {
		t.Fatalf("Expected no error got %v", err)
	}
	if len(encrypted) != 2048 {
		t.Fatalf("Encrypted data length expected \"%v\" got %v", 2048, len(encrypted))
	}
}

func TestEncryptDataLengthEqualsPadding(t *testing.T) {
	enc := New(testKey, 4096, uint32(0), hashFunc)

	data := make([]byte, 4096)

	encrypted, err := enc.Encrypt(data)
	if err != nil {
		t.Fatalf("Expected no error got %v", err)
	}
	encryptedHex := common.Bytes2Hex(encrypted)
	expectedTransformed := common.Hex2Bytes(expectedTransformedHex)

	if !bytes.Equal(encrypted, expectedTransformed) {
		t.Fatalf("Expected %v got %v", expectedTransformedHex, encryptedHex)
	}
}

func TestEncryptDataLengthSmallerThanPadding(t *testing.T) {
	enc := New(testKey, 4096, uint32(0), hashFunc)

	data := make([]byte, 4080)

	encrypted, err := enc.Encrypt(data)
	if err != nil {
		t.Fatalf("Expected no error got %v", err)
	}
	if len(encrypted) != 4096 {
		t.Fatalf("Encrypted data length expected %v got %v", 4096, len(encrypted))
	}
}

func TestDecryptDataLengthNotEqualsPadding(t *testing.T) {
	enc := New(testKey, 4096, uint32(0), hashFunc)

	data := make([]byte, 4097)

	expectedError := "Data length different than padding, data length 4097 padding 4096"

	_, err := enc.Decrypt(data)
	if err == nil || err.Error() != expectedError {
		t.Fatalf("Expected error \"%v\" got \"%v\"", expectedError, err)
	}
}

func TestEncryptDecryptIsIdentity(t *testing.T) {
	testEncryptDecryptIsIdentity(t, 2048, 0, 2048, 32)
	testEncryptDecryptIsIdentity(t, 4096, 0, 4096, 32)
	testEncryptDecryptIsIdentity(t, 4096, 0, 1000, 32)
	testEncryptDecryptIsIdentity(t, 32, 10, 32, 32)
}

func generateRandomData(l int) (r io.Reader, slice []byte) {
	slice = make([]byte, l)
	rand.Seed(time.Now().Unix())
	if _, err := rand.Read(slice); err != nil {
		panic("rand error")
	}
	r = io.LimitReader(bytes.NewReader(slice), int64(l))
	return
}

func testEncryptDecryptIsIdentity(t *testing.T, padding int, initCtr uint32, dataLength int, keyLength int) {
	key := GenerateRandomKey(keyLength)
	enc := New(key, padding, initCtr, hashFunc)

	_, data := generateRandomData(dataLength)

	encrypted, err := enc.Encrypt(data)
	if err != nil {
		t.Fatalf("Expected no error got %v", err)
	}

	decrypted, err := enc.Decrypt(encrypted)
	if err != nil {
		t.Fatalf("Expected no error got %v", err)
	}
	if len(decrypted) != padding {
		t.Fatalf("Expected decrypted data length %v got %v", padding, len(decrypted))
	}

	// we have to remove the extra bytes which were randomly added to fill until padding
	if len(data) < padding {
		decrypted = decrypted[:len(data)]
	}

	if !bytes.Equal(data, decrypted) {
		t.Fatalf("Expected decrypted %v got %v", common.Bytes2Hex(data), common.Bytes2Hex(decrypted))
	}
}
