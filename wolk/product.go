// Copyright 2018 Wolk Inc.
// This file is part of the Wolk library.
package wolk

import (
	"bytes"
	"encoding/json"

	"github.com/ethereum/go-ethereum/common"
)

/*
Applications are bookmarks into developer buckets, stamped with a blockNumber.

Developers register their products in a "products" collection.  Each of the documents there should be of type Product, with this reused mapping
    AddProduct()  => SetKey; DeleteProduct() => DeleteKey; GetProduct() => GetKey; GetProducts() => ScanCollection

Users install apps in a "apps" collection.  Each of the documents there are of type ProductInstance
    InstallApp() => SetKey;  UninstallApp() => DeleteKey;  GetApp() => GetKey;     GetApps() => ScanCollection

The linkage between the two is:
 1. BucketItem.Key <-> Product.Url
 2. BucketItem.ValHash => ProductInstance

The browser must use the ProductInstance.Permissions data to control access to the user data: Permission + BucketNames

When using an application "Docify", the browser fetches the ProductInstance and gets the PermissionsList.

wolk.js within the browser uses the PermissionsList as a gate as to what can and cannot be done.

PermissionsList is disclosed to the user -- complex applications will access different user buckets.

System Buckets are explicitly reserved, but otherwise products register names with "setName" that go into the same space as users.
People reserving their wolk://username bucket is the same as a developer reserving the wolk://docify bucket for their docify application.
Then, system buckets like "friends" are kept in the same name SMT, and users keep their system buckets like

  wolk://username/{contacts,friends,docs,events, ...}
  records are https://en.wikipedia.org/wiki/JSON-LD

For these system buckets, the data schema that they conform to is readable by all and records within the collections conform to that schema.

 wolk://contacts/schema/1.0
 wolk://contacts/schema/1.1
 wolk://contacts/schema/2.0

The Collection definition should refer to the schema, in a manner to be determined -- we could use 	 http://xmlns.com/foaf/spec/ / Google Knowledge Graph

Application developers should/must specify the schema of the data that they write rather than develop proprietary formats.

*/

type Product struct {
	ProductType     uint8  `json:"bucketType"` // 0=free application
	Name            string `json:"name"`       // docify
	URL             string `json:"url"`        // wolk://author/docify
	CreateTime      uint64 `json:"CreateTime"`
	UpdateTime      uint64 `json:"UpdateTime"`
	PermissionsList []Permission
	Writers         []common.Address `json:"Writers"`
}

// users have product instances in their system bucket
type ProductInstance struct {
	Name            string
	PermissionsList []Permission // r+w permissions (e.g. GetKey, ScanCollection, ReadSQL for decrypting data) ABOUT buckets ("photos")
}

type Permission struct {
	PermissionLevel string // r, r+w
	BucketName      string
}

const (
	ProductApp = 0
)

func NewProduct(productType uint8, name string) *Product {
	return &Product{
		ProductType: productType,
		Name:        name,
	}
}

func (p *Product) ValidWriter(addr common.Address, productOwner common.Address) bool {
	if bytes.Compare(addr.Bytes(), productOwner.Bytes()) == 0 {
		return true
	}
	for _, w := range p.Writers {
		if bytes.Compare(w.Bytes(), addr.Bytes()) == 0 {
			return true
		}
	}
	return false
}

func (p *Product) String() string {
	bytes, err := json.Marshal(p)
	if err != nil {
		return "{}"
	} else {
		return string(bytes)
	}
}

func ParseProductList(rawProductList []byte) (products []*Product, err error) {
	err = json.Unmarshal(rawProductList, &products)
	if err != nil {
		return products, err
	}
	return products, nil
}
