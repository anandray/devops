const contract_abi = [
    {
       'constant': true,
       'inputs': [
          {
             'name': '',
             'type': 'uint64'
          }
       ],
       'name': 'depositIndex',
       'outputs': [
          {
             'name': '',
             'type': 'uint64'
          }
       ],
       'payable': false,
       'stateMutability': 'view',
       'type': 'function'
    },
    {
       'constant': true,
       'inputs': [
       ],
       'name': 'getNextExit',
       'outputs': [
          {
             'name': 'depID',
             'type': 'uint64'
          },
          {
             'name': 'tokenID',
             'type': 'uint64'
          },
          {
             'name': 'exitableTS',
             'type': 'uint256'
          }
       ],
       'payable': false,
       'stateMutability': 'view',
       'type': 'function'
    },
    {
       'constant': false,
       'inputs': [
          {
             'name': 'txBytes',
             'type': 'bytes'
          },
          {
             'name': 'proof',
             'type': 'bytes'
          },
          {
             'name': 'blk',
             'type': 'uint64'
          }
       ],
       'name': 'challenge',
       'outputs': [
       ],
       'payable': false,
       'stateMutability': 'nonpayable',
       'type': 'function'
    },
    {
       'constant': true,
       'inputs': [
       ],
       'name': 'currentDepositIndex',
       'outputs': [
          {
             'name': '',
             'type': 'uint64'
          }
       ],
       'payable': false,
       'stateMutability': 'view',
       'type': 'function'
    },
    {
       'constant': false,
       'inputs': [
          {
             'name': '_depositIndex',
             'type': 'uint64'
          }
       ],
       'name': 'depositExit',
       'outputs': [
       ],
       'payable': false,
       'stateMutability': 'nonpayable',
       'type': 'function'
    },
    {
       'constant': false,
       'inputs': [
       ],
       'name': 'kill',
       'outputs': [
       ],
       'payable': false,
       'stateMutability': 'nonpayable',
       'type': 'function'
    },
    {
       'constant': true,
       'inputs': [
          {
             'name': '',
             'type': 'uint64'
          }
       ],
       'name': 'childChain',
       'outputs': [
          {
             'name': '',
             'type': 'bytes32'
          }
       ],
       'payable': false,
       'stateMutability': 'view',
       'type': 'function'
    },
    {
       'constant': false,
       'inputs': [
          {
             'name': 'prevTxBytes',
             'type': 'bytes'
          },
          {
             'name': 'prevProof',
             'type': 'bytes'
          },
          {
             'name': 'prevBlk',
             'type': 'uint64'
          },
          {
             'name': 'txBytes',
             'type': 'bytes'
          },
          {
             'name': 'proof',
             'type': 'bytes'
          },
          {
             'name': 'blk',
             'type': 'uint64'
          }
       ],
       'name': 'startExit',
       'outputs': [
       ],
       'payable': false,
       'stateMutability': 'nonpayable',
       'type': 'function'
    },
    {
       'constant': false,
       'inputs': [
          {
             'name': 'txBytes',
             'type': 'bytes'
          },
          {
             'name': 'proof',
             'type': 'bytes'
          },
          {
             'name': 'blk',
             'type': 'uint64'
          },
          {
             'name': 'faultyTxBytes',
             'type': 'bytes'
          },
          {
             'name': 'faultyProof',
             'type': 'bytes'
          },
          {
             'name': 'faultyBlk',
             'type': 'uint64'
          }
       ],
       'name': 'challengeBefore',
       'outputs': [
       ],
       'payable': false,
       'stateMutability': 'nonpayable',
       'type': 'function'
    },
    {
       'constant': true,
       'inputs': [
       ],
       'name': 'currentBlkNum',
       'outputs': [
          {
             'name': '',
             'type': 'uint64'
          }
       ],
       'payable': false,
       'stateMutability': 'view',
       'type': 'function'
    },
    {
       'constant': true,
       'inputs': [
       ],
       'name': 'authority',
       'outputs': [
          {
             'name': '',
             'type': 'address'
          }
       ],
       'payable': false,
       'stateMutability': 'view',
       'type': 'function'
    },
    {
       'constant': false,
       'inputs': [
       ],
       'name': 'finalizeExits',
       'outputs': [
       ],
       'payable': false,
       'stateMutability': 'nonpayable',
       'type': 'function'
    },
    {
       'constant': false,
       'inputs': [
       ],
       'name': 'deposit',
       'outputs': [
       ],
       'payable': true,
       'stateMutability': 'payable',
       'type': 'function'
    },
    {
       'constant': true,
       'inputs': [
          {
             'name': '',
             'type': 'uint64'
          }
       ],
       'name': 'exits',
       'outputs': [
          {
             'name': 'prevBlk',
             'type': 'uint64'
          },
          {
             'name': 'exitBlk',
             'type': 'uint64'
          },
          {
             'name': 'exitor',
             'type': 'address'
          },
          {
             'name': 'exitableTS',
             'type': 'uint256'
          },
          {
             'name': 'bond',
             'type': 'uint256'
          }
       ],
       'payable': false,
       'stateMutability': 'view',
       'type': 'function'
    },
    {
       'constant': false,
       'inputs': [
          {
             'name': '_blkRoot',
             'type': 'bytes32'
          },
          {
             'name': '_blknum',
             'type': 'uint64'
          }
       ],
       'name': 'submitBlock',
       'outputs': [
       ],
       'payable': false,
       'stateMutability': 'nonpayable',
       'type': 'function'
    },
    {
       'constant': true,
       'inputs': [
          {
             'name': '',
             'type': 'uint64'
          }
       ],
       'name': 'depositBalance',
       'outputs': [
          {
             'name': '',
             'type': 'uint64'
          }
       ],
       'payable': false,
       'stateMutability': 'view',
       'type': 'function'
    },
    {
       'inputs': [
       ],
       'payable': false,
       'stateMutability': 'nonpayable',
       'type': 'constructor'
    },
    {
       'payable': true,
       'stateMutability': 'payable',
       'type': 'fallback'
    },
    {
       'anonymous': false,
       'inputs': [
          {
             'indexed': false,
             'name': '_depositor',
             'type': 'address'
          },
          {
             'indexed': true,
             'name': '_depositIndex',
             'type': 'uint64'
          },
          {
             'indexed': false,
             'name': '_denomination',
             'type': 'uint64'
          },
          {
             'indexed': true,
             'name': '_tokenID',
             'type': 'uint64'
          }
       ],
       'name': 'Deposit',
       'type': 'event'
    },
    {
       'anonymous': false,
       'inputs': [
          {
             'indexed': false,
             'name': '_exiter',
             'type': 'address'
          },
          {
             'indexed': true,
             'name': '_depositIndex',
             'type': 'uint64'
          },
          {
             'indexed': false,
             'name': '_denomination',
             'type': 'uint64'
          },
          {
             'indexed': true,
             'name': '_tokenID',
             'type': 'uint64'
          },
          {
             'indexed': true,
             'name': '_timestamp',
             'type': 'uint256'
          }
       ],
       'name': 'StartExit',
       'type': 'event'
    },
    {
       'anonymous': false,
       'inputs': [
          {
             'indexed': false,
             'name': '_rootHash',
             'type': 'bytes32'
          },
          {
             'indexed': true,
             'name': '_blknum',
             'type': 'uint64'
          },
          {
             'indexed': true,
             'name': '_currentDepositIndex',
             'type': 'uint64'
          }
       ],
       'name': 'PublishedBlock',
       'type': 'event'
    },
    {
       'anonymous': false,
       'inputs': [
          {
             'indexed': false,
             'name': '_exiter',
             'type': 'address'
          },
          {
             'indexed': true,
             'name': '_depositIndex',
             'type': 'uint64'
          },
          {
             'indexed': false,
             'name': '_denomination',
             'type': 'uint64'
          },
          {
             'indexed': true,
             'name': '_tokenID',
             'type': 'uint64'
          },
          {
             'indexed': true,
             'name': '_timestamp',
             'type': 'uint256'
          }
       ],
       'name': 'FinalizedExit',
       'type': 'event'
    },
    {
       'anonymous': false,
       'inputs': [
          {
             'indexed': false,
             'name': '_challenger',
             'type': 'address'
          },
          {
             'indexed': true,
             'name': '_tokenID',
             'type': 'uint64'
          },
          {
             'indexed': true,
             'name': '_timestamp',
             'type': 'uint256'
          }
       ],
       'name': 'Challenge',
       'type': 'event'
    },
    {
       'anonymous': false,
       'inputs': [
          {
             'indexed': true,
             'name': '_priority',
             'type': 'uint256'
          }
       ],
       'name': 'ExitStarted',
       'type': 'event'
    },
    {
       'anonymous': false,
       'inputs': [
          {
             'indexed': true,
             'name': '_priority',
             'type': 'uint256'
          }
       ],
       'name': 'ExitCompleted',
       'type': 'event'
    },
    {
       'anonymous': false,
       'inputs': [
          {
             'indexed': false,
             'name': 'depID',
             'type': 'uint64'
          },
          {
             'indexed': false,
             'name': 'tokenID',
             'type': 'uint64'
          },
          {
             'indexed': false,
             'name': 'exitableTS',
             'type': 'uint256'
          }
       ],
       'name': 'CurrtentExit',
       'type': 'event'
    },
    {
       'anonymous': false,
       'inputs': [
          {
             'indexed': false,
             'name': 'exitableTS',
             'type': 'uint256'
          },
          {
             'indexed': false,
             'name': 'cuurrentTS',
             'type': 'uint256'
          }
       ],
       'name': 'ExitTime',
       'type': 'event'
    }
];

const contract_address = '0x97f33D99d6d473CB938abB893aBd49A2BB1404bf';
// const ws_provider = 'wss://rinkeby.infura.io/ws';
const provider = 'https://rinkeby.infura.io/pJJrBQxSLPzuFF8KGmi0';
const plasma_provider = 'http://plasma.wolk.com:32003';

const Web3 = require('web3');
const WOLK = require('wolkjs');
const EthereumTx = require('ethereumjs-tx');
const ethUtil = require('ethereumjs-util');
const rlp = require('rlp');
const child_process = require("child_process");
const request = require('request');

const web3 = new Web3();
web3.setProvider(new Web3.providers.HttpProvider(provider));

const plasma = new WOLK(plasma_provider);


const contract = web3.eth.contract(contract_abi).at(contract_address);

var currentBlockNumber = web3.eth.blockNumber;
// console.log("currentBlockNumber: " + currentBlockNumber);


const faucetNonce = '0x' + Number(web3.eth.getTransactionCount("0xeDafCC405196A51AE462442E870d0Ea9d5D395A1")).toString(16);
// console.log("faucetNonce: " + faucetNonce); 

const faucetPrivateKey = Buffer.from('e8ab3b2096705ca9ec5de3b32f6e6dc63fac4485e684bc9bc68758a01f2a3fcb', 'hex');

const sendETHTxParams = {
	nonce: faucetNonce,
	gasPrice: '0x098bca5a00',
	gasLimit: '0x5208',
	to: process.argv[2],
	value: '0x354a6ba7a18000', //0.015ETH
	data: '0x',
	chainId: 4
};

const sendETHTx = new EthereumTx(sendETHTxParams);
sendETHTx.sign(faucetPrivateKey);
const serializedSendETHTx = sendETHTx.serialize();


// ======  faucet send ETH START  ======== //
web3.eth.sendRawTransaction('0x' + serializedSendETHTx.toString('hex'), function(err1, sendETHHash) {
	if (!err1) {
		// console.log("sendETHTx Hash:" + sendETHHash);
    // Sleep for 15 seconds, wait for sendETHTx to confirm
    child_process.execSync("sleep 15");
    const sendETHTx15 = web3.eth.getTransaction(sendETHHash);
    // console.log(sendETHTx15);

    
    // ======  deposit to plasma contract START  ======== //
    const accountNonce = '0x' + Number(web3.eth.getTransactionCount(process.argv[2])).toString(16);

    const accountPkBuffer = Buffer.from(process.argv[3], 'hex');

    const depositRawTxParams = {
      nonce: accountNonce,
      gasPrice: '0x98bca5a00', // 41Gwei
      gasLimit: '0x1175d', // 71517
      to: contract_address,
      value: 10000000000000000, // 0.01ETH
      data: '0x',
      chainId: 4
    };

    const depositTx = new EthereumTx(depositRawTxParams);
    depositTx.sign(accountPkBuffer);

    const serializedDepositTx = depositTx.serialize();

    web3.eth.sendRawTransaction('0x' + serializedDepositTx.toString('hex'), function(err2, sendDepositHash) {
      if (!err2) {
        // console.log("depositTx Hash:" + sendDepositHash);
        // Sleep for another 15 seconds, wait for depositTx to confirm
        child_process.execSync("sleep 15");
        const depositTx15 = web3.eth.getTransaction(sendDepositHash);
        // console.log(depositTx15);

        // ======  get deposit event log from plasma contract ======== //
        request.post({
          url: provider,
          form: JSON.stringify({
            "jsonrpc": "2.0",
            "method": "eth_getLogs",
            "params": [
              {
                "topics": ["0x96d929db0520d785b9981429377486f9182e32225c9a8b9f1b371519644cc68a"],
                "address": contract_address,
                "fromBlock": "0x" + currentBlockNumber.toString(16),
                "toBlock": "latest"
              }
            ],
            "id": 4 
          }),
          headers: {
            'Content-Type': 'application/json'
          }
        }, (error, response)=>{
            if(!error && response.statusCode == 200){
                let result = JSON.parse(response.body).result;
                // console.log(result); 

                for (let i = 0; i < result.length; i++) { 
                  let event = result[i];

                  // ======  make sure deposit from the input address  ======== //
                  if (event.data.includes(process.argv[2].substr(2).toLowerCase())) {

                    // ======  get blockchain ID  ======== //
                    const deposit = {};
                    deposit['tokenID'] = '0x' + ethUtil.unpad(Buffer.from(event.topics[2].substr(2), 'hex').toString('hex'));
                    deposit['depositIndex'] = web3.toHex(web3.toDecimal(event.topics[1]));
                    deposit['denomination'] = '0x2386f26fc10000';
                    deposit['depositor'] = process.argv[2];

                    // console.log(deposit);

                    plasma.getBlockchainID(deposit.tokenID, deposit.depositIndex, function(err2, blockchainid) {
                      if (err2) {
                        console.log(err2.message);
                      }
                      else {
                        // console.log("blockchainID: " + ethUtil.addHexPrefix(ethUtil.setLengthLeft(blockchainid, 8).toString('hex')));
                        const blockchainID = ethUtil.addHexPrefix(ethUtil.setLengthLeft(blockchainid, 8).toString('hex'));

                        // ======  send anchor tx  ======== //
                        const unsignedAnchorTx = rlp.encode(
                          [blockchainID, '0x', '0x03c85f1da84d9c6313e0c34bcb5ace945a9b12105988895252b88ce5b769f82b',
                          [ [process.argv[2], '0xA45b77a98E2B840617e2eC6ddfBf71403bdCb683'], [] ], '']
                        );
                        const msgHash = ethUtil.keccak256(unsignedAnchorTx);
                        let signature = ethUtil.ecsign(msgHash, accountPkBuffer);

                        if (signature.v === 28 || signature.v === 1) {
                            signature = Buffer.concat([signature.r, signature.s, Buffer.from([0x01])]);
                        } else if (signature.v === 27 || signature.v === 0) {
                            signature = Buffer.concat([signature.r, signature.s, Buffer.from([0x00])]);
                        }

                        const anchorTxn = {};
                        anchorTxn['blockchainID'] = blockchainID;
                        anchorTxn['blocknumber'] = '0x0';
                        anchorTxn['blockhash'] = '0x03c85f1da84d9c6313e0c34bcb5ace945a9b12105988895252b88ce5b769f82b';
                        anchorTxn['extra'] = {'addedOwners': [process.argv[2], '0xA45b77a98E2B840617e2eC6ddfBf71403bdCb683'],
                          'removedOwners': []};
                        anchorTxn['sig'] = '0x' + signature.toString('hex');
                        // console.log(anchorTxn);

                        
                        plasma.sendAnchorTransaction(anchorTxn, function(err, anchorTxHash) {
                          if (err) {
                            console.log(err.message);
                          }
                          else {
                            // console.log("anchorTxHash: " + anchorTxHash);
                            console.log(blockchainID);
                          }
                        });
                    
                      }
                    });
                  }
                }

            }
        });

      }

    });

	}
});


