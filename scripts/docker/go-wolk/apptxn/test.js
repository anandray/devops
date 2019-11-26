const SWARMDB = require('./index');
const swarmdb = new SWARMDB("http://10.128.0.2:8545");


const owner = "new2.eth";
const dbname = "newdb2";
const tablename = "newcontacts2";

var columns = [
    { "indextype": "BPLUS", "columnname": "email", "columntype": "STRING", "primary": 1 },
    { "indextype": "BPLUS", "columnname": "age", "columntype": "INTEGER", "primary": 0 }
];


/* Create database *
swarmdb.createDatabase(owner, dbname, 1)
.then(console.log);
*/


/* Create tables *
swarmdb.createTable(owner, dbname, tablename, columns)
.then(console.log); 
*/


/* List databases */
swarmdb.listDatabases(owner)
.then(console.log);


/* List tables *
swarmdb.listTables(owner, dbname)
.then(console.log);
*/



/* Describe table *
swarmdb.describeTable(owner, dbname, tablename)
.then(console.log);
*/

/* Put *
//swarmdb.put(owner, dbname, tablename, [ {"age":1,"email":"test05@wolk.com"}, {"age":2,"email":"test06@wolk.com"} ])
swarmdb.put(owner, dbname, tablename, [ {"age":8,"email":"test12@wolk.com"} ])
.then(console.log);
*/

/* Get *
swarmdb.get(owner, dbname, tablename, "test05@wolk.com")
.then(console.log);
*/


/* @ Select Query  
swarmdb.selectQuery(owner, dbname, "select email, age from newcontacts1 where email = 'test003@wolk.com'")
.then(console.log);
*/

/* Insert Query
swarmdb.insertQuery(owner, dbname, "INSERT INTO newcontacts1(email, age) VALUES('test003@wolk.com',6)")
.then(console.log);
 */

/* @ Update Query 
swarmdb.updateQuery(owner, dbname, "UPDATE newcontacts1 SET age=8 WHERE email='test003@wolk.com'")
.then(console.log);
*/

/* Drop table *
swarmdb.dropTable(owner, dbname, tablename)
.then(console.log);
*/

/* @ Drop database *
swarmdb.dropDatabase(owner, dbname)
.then(console.log);
*/


