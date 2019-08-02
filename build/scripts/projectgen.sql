DROP TABLE IF EXISTS project4;
CREATE TABLE `project4` (
    `projectID` varchar(40) NOT NULL,
    `node` int(11) NOT NULL,
    `cloudprovider` varchar(32) NOT NULL,
    `dataprovider` varchar(32) DEFAULT '',
    `instanceGroup` varchar(64) DEFAULT '',
    `region` varchar(32) DEFAULT NULL,
    `active` tinyint(4) DEFAULT '1',
    `lb` varchar(128) DEFAULT NULL,
    `lburl` varchar(1024) DEFAULT NULL,
    `primaryzone` varchar(128) DEFAULT NULL,
    `zones` varchar(255) DEFAULT NULL,
    `MicrosoftAzureAccountName` varchar(32) DEFAULT NULL,
    `MicrosoftAzureAccountKey` varchar(128) DEFAULT NULL,
    `AlibabaAccessKeyId` varchar(64) DEFAULT NULL,
    `AlibabaAccessKeySecret` varchar(64) DEFAULT NULL,
    `AlibabaEndpointURL` varchar(128) DEFAULT NULL,
    PRIMARY KEY (`node`)
  );
insert into project4 (projectID, node, cloudprovider, region) values ('us-east4-wlk', '32', 'gc', 'us-east4');
insert into project4 (projectID, node, cloudprovider, region) values ('australia-southeast1-wlk', '33', 'gc', 'australia-southeast1');
insert into project4 (projectID, node, cloudprovider, region) values ('us-west2-wlk', '34', 'gc', 'us-west2');
insert into project4 (projectID, node, cloudprovider, region) values ('asia-northeast1-wlk', '35', 'gc', 'asia-northeast1');
insert into project4 (projectID, node, cloudprovider, region) values ('europe-west6-wlk', '36', 'gc', 'europe-west6');
insert into project4 (projectID, node, cloudprovider, region) values ('europe-west3-wlk', '37', 'gc', 'europe-west3');
insert into project4 (projectID, node, cloudprovider, region) values ('us-east1-wlk', '38', 'gc', 'us-east1');
insert into project4 (projectID, node, cloudprovider, region) values ('northamerica-northeast1-wlk', '39', 'gc', 'northamerica-northeast1');


DROP TABLE IF EXISTS project5;
CREATE TABLE `project5` (
    `projectID` varchar(40) NOT NULL,
    `node` int(11) NOT NULL,
    `cloudprovider` varchar(32) NOT NULL,
    `dataprovider` varchar(32) DEFAULT '',
    `instanceGroup` varchar(64) DEFAULT '',
    `region` varchar(32) DEFAULT NULL,
    `active` tinyint(4) DEFAULT '1',
    `lb` varchar(128) DEFAULT NULL,
    `lburl` varchar(1024) DEFAULT NULL,
    `primaryzone` varchar(128) DEFAULT NULL,
    `zones` varchar(255) DEFAULT NULL,
    `MicrosoftAzureAccountName` varchar(32) DEFAULT NULL,
    `MicrosoftAzureAccountKey` varchar(128) DEFAULT NULL,
    `AlibabaAccessKeyId` varchar(64) DEFAULT NULL,
    `AlibabaAccessKeySecret` varchar(64) DEFAULT NULL,
    `AlibabaEndpointURL` varchar(128) DEFAULT NULL,
    PRIMARY KEY (`node`)
  );
insert into project5 (projectID, node, cloudprovider, region) values ('asia-south1-wlk', '40', 'gc', 'asia-south1');
insert into project5 (projectID, node, cloudprovider, region) values ('europe-west3-wlk', '41', 'gc', 'europe-west3');
insert into project5 (projectID, node, cloudprovider, region) values ('us-east4-wlk', '42', 'gc', 'us-east4');
insert into project5 (projectID, node, cloudprovider, region) values ('southamerica-east1-wlk', '43', 'gc', 'southamerica-east1');
insert into project5 (projectID, node, cloudprovider, region) values ('europe-west2-wlk', '44', 'gc', 'europe-west2');
insert into project5 (projectID, node, cloudprovider, region) values ('europe-west6-wlk', '45', 'gc', 'europe-west6');
insert into project5 (projectID, node, cloudprovider, region) values ('asia-northeast1-wlk', '46', 'gc', 'asia-northeast1');
insert into project5 (projectID, node, cloudprovider, region) values ('us-east1-wlk', '47', 'gc', 'us-east1');


DROP TABLE IF EXISTS project6;
CREATE TABLE `project6` (
    `projectID` varchar(40) NOT NULL,
    `node` int(11) NOT NULL,
    `cloudprovider` varchar(32) NOT NULL,
    `dataprovider` varchar(32) DEFAULT '',
    `instanceGroup` varchar(64) DEFAULT '',
    `region` varchar(32) DEFAULT NULL,
    `active` tinyint(4) DEFAULT '1',
    `lb` varchar(128) DEFAULT NULL,
    `lburl` varchar(1024) DEFAULT NULL,
    `primaryzone` varchar(128) DEFAULT NULL,
    `zones` varchar(255) DEFAULT NULL,
    `MicrosoftAzureAccountName` varchar(32) DEFAULT NULL,
    `MicrosoftAzureAccountKey` varchar(128) DEFAULT NULL,
    `AlibabaAccessKeyId` varchar(64) DEFAULT NULL,
    `AlibabaAccessKeySecret` varchar(64) DEFAULT NULL,
    `AlibabaEndpointURL` varchar(128) DEFAULT NULL,
    PRIMARY KEY (`node`)
  );
insert into project6 (projectID, node, cloudprovider, region) values ('wolk-48-azure-ukwest', '48', 'azure', 'ukwest');
insert into project6 (projectID, node, cloudprovider, region) values ('europe-west6-wlk', '49', 'gc', 'europe-west6');
insert into project6 (projectID, node, cloudprovider, region) values ('wolk-50-azure-canadacentral', '50', 'azure', 'canadacentral');
insert into project6 (projectID, node, cloudprovider, region) values ('asia-northeast1-wlk', '51', 'gc', 'asia-northeast1');
insert into project6 (projectID, node, cloudprovider, region) values ('us-west2-wlk', '52', 'gc', 'us-west2');
insert into project6 (projectID, node, cloudprovider, region) values ('wolk-53-azure-westindia', '53', 'azure', 'westindia');
insert into project6 (projectID, node, cloudprovider, region) values ('us-east4-wlk', '54', 'gc', 'us-east4');
insert into project6 (projectID, node, cloudprovider, region) values ('wolk-55-azure-westus', '55', 'azure', 'westus');


DROP TABLE IF EXISTS project7;
CREATE TABLE `project7` (
    `projectID` varchar(40) NOT NULL,
    `node` int(11) NOT NULL,
    `cloudprovider` varchar(32) NOT NULL,
    `dataprovider` varchar(32) DEFAULT '',
    `instanceGroup` varchar(64) DEFAULT '',
    `region` varchar(32) DEFAULT NULL,
    `active` tinyint(4) DEFAULT '1',
    `lb` varchar(128) DEFAULT NULL,
    `lburl` varchar(1024) DEFAULT NULL,
    `primaryzone` varchar(128) DEFAULT NULL,
    `zones` varchar(255) DEFAULT NULL,
    `MicrosoftAzureAccountName` varchar(32) DEFAULT NULL,
    `MicrosoftAzureAccountKey` varchar(128) DEFAULT NULL,
    `AlibabaAccessKeyId` varchar(64) DEFAULT NULL,
    `AlibabaAccessKeySecret` varchar(64) DEFAULT NULL,
    `AlibabaEndpointURL` varchar(128) DEFAULT NULL,
    PRIMARY KEY (`node`)
  );
insert into project7 (projectID, node, cloudprovider, region) values ('wolk-56-azure-australiasoutheast', '56', 'azure', 'australiasoutheast');
insert into project7 (projectID, node, cloudprovider, region) values ('wolk-57-azure-westus2', '57', 'azure', 'westus2');
insert into project7 (projectID, node, cloudprovider, region) values ('wolk-58-azure-eastus', '58', 'azure', 'eastus');
insert into project7 (projectID, node, cloudprovider, region) values ('wolk-59-azure-southafricanorth', '59', 'azure', 'southafricanorth');
insert into project7 (projectID, node, cloudprovider, region) values ('wolk-60-azure-canadacentral', '60', 'azure', 'canadacentral');
insert into project7 (projectID, node, cloudprovider, region) values ('wolk-61-azure-eastus2', '61', 'azure', 'eastus2');
insert into project7 (projectID, node, cloudprovider, region) values ('australia-southeast1-wlk', '62', 'gc', 'australia-southeast1');
insert into project7 (projectID, node, cloudprovider, region) values ('northamerica-northeast1-wlk', '63', 'gc', 'northamerica-northeast1');


DROP TABLE IF EXISTS project8;
CREATE TABLE `project8` (
    `projectID` varchar(40) NOT NULL,
    `node` int(11) NOT NULL,
    `cloudprovider` varchar(32) NOT NULL,
    `dataprovider` varchar(32) DEFAULT '',
    `instanceGroup` varchar(64) DEFAULT '',
    `region` varchar(32) DEFAULT NULL,
    `active` tinyint(4) DEFAULT '1',
    `lb` varchar(128) DEFAULT NULL,
    `lburl` varchar(1024) DEFAULT NULL,
    `primaryzone` varchar(128) DEFAULT NULL,
    `zones` varchar(255) DEFAULT NULL,
    `MicrosoftAzureAccountName` varchar(32) DEFAULT NULL,
    `MicrosoftAzureAccountKey` varchar(128) DEFAULT NULL,
    `AlibabaAccessKeyId` varchar(64) DEFAULT NULL,
    `AlibabaAccessKeySecret` varchar(64) DEFAULT NULL,
    `AlibabaEndpointURL` varchar(128) DEFAULT NULL,
    PRIMARY KEY (`node`)
  );
insert into project8 (projectID, node, cloudprovider, region) values ('us-east4-wlk', '64', 'gc', 'us-east4');
insert into project8 (projectID, node, cloudprovider, region) values ('asia-northeast1-wlk', '65', 'gc', 'asia-northeast1');
insert into project8 (projectID, node, cloudprovider, region) values ('australia-southeast1-wlk', '66', 'gc', 'australia-southeast1');
insert into project8 (projectID, node, cloudprovider, region) values ('wolk-67-azure-eastus', '67', 'azure', 'eastus');
insert into project8 (projectID, node, cloudprovider, region) values ('asia-south1-wlk', '68', 'gc', 'asia-south1');
insert into project8 (projectID, node, cloudprovider, region) values ('wolk-69-azure-westeurope', '69', 'azure', 'westeurope');
insert into project8 (projectID, node, cloudprovider, region) values ('wolk-70-azure-westus', '70', 'azure', 'westus');
insert into project8 (projectID, node, cloudprovider, region) values ('wolk-71-azure-canadaeast', '71', 'azure', 'canadaeast');


DROP TABLE IF EXISTS project9;
CREATE TABLE `project9` (
    `projectID` varchar(40) NOT NULL,
    `node` int(11) NOT NULL,
    `cloudprovider` varchar(32) NOT NULL,
    `dataprovider` varchar(32) DEFAULT '',
    `instanceGroup` varchar(64) DEFAULT '',
    `region` varchar(32) DEFAULT NULL,
    `active` tinyint(4) DEFAULT '1',
    `lb` varchar(128) DEFAULT NULL,
    `lburl` varchar(1024) DEFAULT NULL,
    `primaryzone` varchar(128) DEFAULT NULL,
    `zones` varchar(255) DEFAULT NULL,
    `MicrosoftAzureAccountName` varchar(32) DEFAULT NULL,
    `MicrosoftAzureAccountKey` varchar(128) DEFAULT NULL,
    `AlibabaAccessKeyId` varchar(64) DEFAULT NULL,
    `AlibabaAccessKeySecret` varchar(64) DEFAULT NULL,
    `AlibabaEndpointURL` varchar(128) DEFAULT NULL,
    PRIMARY KEY (`node`)
  );
insert into project9 (projectID, node, cloudprovider, region) values ('europe-west2-wlk', '72', 'gc', 'europe-west2');
insert into project9 (projectID, node, cloudprovider, region) values ('wolk-73-azure-francecentral', '73', 'azure', 'francecentral');
insert into project9 (projectID, node, cloudprovider, region) values ('wolk-74-azure-southcentralus', '74', 'azure', 'southcentralus');
insert into project9 (projectID, node, cloudprovider, region) values ('asia-northeast1-wlk', '75', 'gc', 'asia-northeast1');
insert into project9 (projectID, node, cloudprovider, region) values ('wolk-76-azure-australiasoutheast', '76', 'azure', 'australiasoutheast');
insert into project9 (projectID, node, cloudprovider, region) values ('wolk-77-azure-southeastasia', '77', 'azure', 'southeastasia');
insert into project9 (projectID, node, cloudprovider, region) values ('wolk-78-azure-westus', '78', 'azure', 'westus');
insert into project9 (projectID, node, cloudprovider, region) values ('wolk-79-azure-canadacentral', '79', 'azure', 'canadacentral');


