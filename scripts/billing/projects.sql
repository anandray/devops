drop table if exists project;
CREATE TABLE `project` (
  `projectID` varchar(40) NOT NULL,
  `node` int NOT NULL,
  `cloudprovider` varchar(32) NOT NULL,
  `dataprovider` varchar(32) NOT NULL,
  `instanceGroup` varchar(64) DEFAULT NULL,
  `gcregion` varchar(32) DEFAULT NULL,
  `active` tinyint(4) DEFAULT '1',
  `lb` varchar(128) DEFAULT NULL,
  `primaryzone` varchar(128) DEFAULT NULL,
  `zones` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`projectID`)
);

-- live on 1/18
insert into project (projectID, node, cloudprovider, dataprovider, instanceGroup, gcregion, lb, primaryzone, zones, active) values ("wolk-us-west", 0, "gc", "google_datastore", "wolk-0-gc-us-west-datastore", "us-west2", "us-west.wolk.com", "us-west2-b", "us-west2-a,us-west2-b,us-west2-c", 1);
insert into project (projectID, node, cloudprovider, dataprovider, instanceGroup, gcregion, lb, primaryzone, zones, active) values ("wolk-us-east", 1, "gc", "google_datastore", "wolk-1-gc-us-east-datastore", "us-east4",  "us-east.wolk.com", "us-east4-b", "us-east4-a,us-east4-b,us-east4-c", 1);
insert into project (projectID, node, cloudprovider, dataprovider, instanceGroup, gcregion, lb, primaryzone, zones, active) values ("wolk-us-central", 2, "gc", "google_datastore", "wolk-2-gc-us-central-datastore", "us-central1",  "us-central.wolk.com", "us-central1-b", "us-central1-a,us-central1-b,us-central1-c,us-central1-f", 1);
insert into project (projectID, node, cloudprovider, dataprovider, instanceGroup, gcregion, lb, primaryzone, zones, active) values ("wolk-asia-east", 3, "gc", "google_datastore", "wolk-3-gc-asia-east-datastore", "asia-east2",  "asia-east.wolk.com", "asia-east2-b", "asia-east2-a,asia-east2-b,asia-east2-c", 1);
insert into project (projectID, node, cloudprovider, dataprovider, instanceGroup, gcregion, lb, primaryzone, zones, active) values ("wolk-europe-west", 4, "gc", "google_datastore", "wolk-4-gc-europe-west-datastore", "europe-west3",  "europe-west3.wolk.com", "europe-west3-b", "europe-west3-a,europe-west3-b,europe-west3-c", 1);
insert into project (projectID, node, cloudprovider, dataprovider, instanceGroup, gcregion, lb, primaryzone, zones, active) values ("wolk-asia-south", 5, "gc", "google_datastore", "wolk-5-gc-asia-south-datastore", "asia-south1",  "asia-south.wolk.com", "asia-south1-b", "asia-south1-a,asia-south1-b,asia-south1-c", 1);
insert into project (projectID, node, cloudprovider, dataprovider, instanceGroup, gcregion, lb, primaryzone, zones, active) values ("wolk-northamerica-northeast", 6, "gc", "google_datastore", "wolk-6-gc-northamerica-northeast-datastore", "northamerica-northeast1",  "northamerica-northeast.wolk.com", "northamerica-northeast1-b", "asia-south1-a,asia-south1-b,asia-south1-c", 1);
insert into project (projectID, node, cloudprovider, dataprovider, instanceGroup, gcregion, lb, primaryzone, zones, active) values ("wolk-southamerica-east", 7, "gc", "google_datastore", "wolk-7-gc-southamerica-east-datastore", "southamerica-east1",  "southamerica-east.wolk.com", "southamerica-east1-b", "southamerica-east1-a,southamerica-east1-b,southamerica-east1-c", 1);

-- to be setup on 1/19
insert into project (projectID, node, cloudprovider, dataprovider, instanceGroup, gcregion, lb, primaryzone, zones, active) values ("australia-southeast", 8, "gc", "google_datastore", "wolk-8-gc-australia-southeast-datastore", "australia-southeast1",  "australia-southeast.wolk.com", "australia-southeast1-a", "australia-southeast1-a,australia-southeast1-b,australia-southeast1-c", 0);
insert into project (projectID, node, cloudprovider, dataprovider, instanceGroup, gcregion, lb, primaryzone, zones, active) values ("asia-northeast", 9, "gc", "google_datastore", "wolk-9-gc-asia-northeast-datastore", "asia-northeast1",  "asia-northeast.wolk.com", "asia-northeast1-a", "asia-northeast1-a,asia-northeast1-b,asia-northeast1-c", 0);
insert into project (projectID, node, cloudprovider, dataprovider, instanceGroup, gcregion, lb, primaryzone, zones, active) values ("europe-west2", 10, "gc", "google_datastore", "wolk-10-gc-europe-west2-datastore", "europe-west2",  "europe-west2.wolk.com", "europe-west2-a", "europe-west2-a,europe-west2-b,europe-west2-c", 0);



