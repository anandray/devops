const Web3 = require('web3');
const key = "0x3d51e84f0270019e9238f6946bd35a8f";
const from = "0x1d21f64b4048d91b8216209fb682c797d63f5dd1";


class SWARMDB {
	constructor(provider) {
	this.web3 = new Web3(); 
        this.web3.setProvider(new Web3.providers.HttpProvider(provider));
    }

    sendAppTransaction(msg) {
	console.log(msg + "\n");
    	let data = this.encrypt(msg, key);
    	let request = {from: from, data: data};
    	return this.web3.eth.sendAppTransaction(request);
    }

    swarmdbCall(msg) {
	let data = this.encrypt(msg, key);
	let request = {from: from, data: data};
    	return this.web3.eth.swarmdbCall(request);
    }

    createDatabase(owner, database, encrypted) {
        let msg = JSON.stringify({
            "requesttype": "CreateDatabase",
            "owner": owner,
            "database": database,
            "encrypted": encrypted
        });
        return this.sendAppTransaction(msg);
    }

	listDatabases(owner) {
        let msg = JSON.stringify({
            "requesttype": "ListDatabases",
            "owner": owner
        });
        return this.swarmdbCall(msg);
    }

    dropDatabase(owner, database) {
        let msg = JSON.stringify({
            "requesttype": "DropDatabase",
            "owner": owner,
            "database": database
        });
        return this.sendAppTransaction(msg);
    }

    createTable(owner, database, table, columns) {
        let msg = JSON.stringify({
            "requesttype": "CreateTable",
            "owner": owner,
            "database": database,
            "table": table,
            "columns": columns
        });
        return this.sendAppTransaction(msg);
    }	

    listTables(owner, database) {
        let msg = JSON.stringify({
            "requesttype": "ListTables",
            "owner": owner,
            "database": database
        });
        return this.swarmdbCall(msg);
    }

	describeTable(owner, database, table) {
        let msg = JSON.stringify({
            "requesttype": "DescribeTable",
            "owner": owner,
            "database": database,
            "table": table
        });
        return this.swarmdbCall(msg);
    }

    dropTable(owner, database, table) {
        let msg = JSON.stringify({
            "requesttype": "DropTable",
            "owner": owner,
            "database": database,
            "table": table
        });
        return this.sendAppTransaction(msg);
    }

    get(owner, database, table, key) {
        let msg = JSON.stringify({
            "requesttype": "Get",
            "owner": owner,
            "database": database,
            "table": table,
            "key": key,
            "columns": null
        });
        return this.swarmdbCall(msg);
    }

    put(owner, database, table, rows) {
        let msg = JSON.stringify({
            "requesttype": "Put",
            "owner": owner,
            "database": database,
            "table": table,
            "rows": rows,
            "columns": null
        });
        return this.sendAppTransaction(msg);
    }

    insertQuery(owner, database, queryStatement) {
        let msg = JSON.stringify({
            "requesttype": "Query",
            "owner": owner,
            "database": database,
            "Query": queryStatement
        });
        return this.sendAppTransaction(msg);
    }

    updateQuery(owner, database, queryStatement) {
        let msg = JSON.stringify({
            "requesttype": "Query",
            "owner": owner,
            "database": database,
            "Query": queryStatement
        });
        return this.sendAppTransaction(msg);
    }                   

    selectQuery(owner, database, queryStatement) {
        let msg = JSON.stringify({
            "requesttype": "Query",
            "owner": owner,
            "database": database,
            "Query": queryStatement
        });
        return this.swarmdbCall(msg);
    }   

    encrypt(data, key, padding = 0, ctr = 0) {
    	var data = Buffer.from(data, 'utf8').toString('hex');
    	data = "0x" + data;

	let length = (data.length / 2) - 1;
	let isFixedPadding = padding > 0;

	if (isFixedPadding && (length > padding)) {
		throw "Data length longer than padding";
	}

	let paddedData = data;
	if (isFixedPadding && (length < padding)) {
		paddedData += rand_string((padding - length) * 2);
	}
	return this.transform(paddedData, key, ctr);
    }

    transform(data, key, ctr = 0) {
	let dataLength = (data.length / 2) - 1;
	let transformedData = "0x" + "0".repeat(dataLength * 2);
	let transformedDataArr = this.web3.utils.hexToBytes(transformedData);
	let dataArr = this.web3.utils.hexToBytes(data);

	let hashSize = 32;
	console.log(data + "<<Data in transform\n")
	for (let i = 0; i < dataLength; i += hashSize) {
			let tmp = key;
			let ctrBytes = this.getCTRBytes(ctr);
			tmp = key + ctrBytes.substring(2);

			let ctrHash = this.web3.utils.keccak256(tmp);

			let segmentKey = this.web3.utils.keccak256(ctrHash);
			let segmentKeyArr = this.web3.utils.hexToBytes(segmentKey);

			let segmentSize = Math.min(hashSize, dataLength - i);

			for (let j = 0; j < segmentSize; j++) {
				transformedDataArr[i + j] = dataArr[i + j] ^ segmentKeyArr[j];
			}

			ctr++;
	}
	console.log(this.web3.utils.bytesToHex(transformedDataArr) + "<<Data POST transform\n")
	return this.web3.utils.bytesToHex(transformedDataArr);
    }

	getCTRBytes(ctr) {
		let hex = this.web3.utils.numberToHex(ctr);
		if (hex.length % 2 != 0) {
			hex = this.web3.utils.padLeft(hex, hex.length - 1);
		}
		return this.web3.utils.padRight(hex, 8);
	}
}

module.exports = SWARMDB;
