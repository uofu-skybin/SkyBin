package metaserver

import (
	"fmt"
	"regexp"
	"skybin/core"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

const dbName = "skybin"
const dbAddress = "127.0.0.1"

type mongoDB struct {
	session *mgo.Session
}

type fileVersion struct {
	Number int
	ID     string
}

func newMongoDB() (*mongoDB, error) {
	session, err := mgo.Dial(dbAddress)
	if err != nil {
		return nil, err
	}

	db := mongoDB{session: session}
	return &db, nil
}

func (db *mongoDB) CloseDB() {
	db.session.Close()
}

func (db *mongoDB) getMongoCollection(name string) (*mgo.Collection, *mgo.Session, error) {
	session := db.session.Copy()
	c := session.DB(dbName).C(name)

	return c, session, nil
}

func (db *mongoDB) findAllFromCollection(collection string, result interface{}) error {
	c, session, err := db.getMongoCollection(collection)
	if err != nil {
		return err
	}
	defer session.Close()

	err = c.Find(nil).All(result)
	if err != nil {
		return err
	}

	return nil
}

func (db *mongoDB) findOneFromCollectionByID(collection string, ID string, result interface{}) error {
	c, session, err := db.getMongoCollection(collection)
	if err != nil {
		return err
	}
	defer session.Close()

	selector := struct{ ID string }{ID: ID}
	err = c.Find(selector).One(result)
	if err != nil {
		return err
	}

	return nil
}

func (db *mongoDB) insertIntoCollection(collection string, doc interface{}) error {
	c, session, err := db.getMongoCollection(collection)
	if err != nil {
		return err
	}
	defer session.Close()

	err = c.Insert(doc)
	if err != nil {
		return err
	}

	return nil
}

func (db *mongoDB) updateInCollection(collection string, ID string, doc interface{}) error {
	c, session, err := db.getMongoCollection(collection)
	if err != nil {
		return err
	}
	defer session.Close()

	selector := struct{ ID string }{ID: ID}
	err = c.Update(selector, doc)
	if err != nil {
		return err
	}

	return nil
}

func (db *mongoDB) deleteInCollectionByID(collection string, ID string) error {
	c, session, err := db.getMongoCollection(collection)
	if err != nil {
		return err
	}
	defer session.Close()

	selector := struct{ ID string }{ID: ID}
	err = c.Remove(selector)
	if err != nil {
		return err
	}

	return nil
}

// Renter operations
//==================

// Return a list of all renters in the database
func (db *mongoDB) FindAllRenters() ([]core.RenterInfo, error) {
	var result []core.RenterInfo
	err := db.findAllFromCollection("renters", &result)
	if err != nil {
		return nil, err
	}
	if result == nil {
		result = make([]core.RenterInfo, 0)
	}
	return result, nil
}

// Return the renter with the specified ID.
func (db *mongoDB) FindRenterByID(renterID string) (*core.RenterInfo, error) {
	var result core.RenterInfo
	err := db.findOneFromCollectionByID("renters", renterID, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Return the renter with the specified alias.
func (db *mongoDB) FindRenterByAlias(alias string) (*core.RenterInfo, error) {
	c, session, err := db.getMongoCollection("renters")
	if err != nil {
		return nil, err
	}
	defer session.Close()

	selector := bson.M{"alias": alias}
	var result *core.RenterInfo
	err = c.Find(selector).One(&result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// Insert the provided renter into the database.
func (db *mongoDB) InsertRenter(renter *core.RenterInfo) error {
	err := db.insertIntoCollection("renters", renter)
	if err != nil {
		return err
	}
	return nil
}

// Update the provided renter in the databse.
func (db *mongoDB) UpdateRenter(renter *core.RenterInfo) error {
	err := db.updateInCollection("renters", renter.ID, renter)
	if err != nil {
		return err
	}
	return nil
}

// Delete the specified renter from the database.
func (db *mongoDB) DeleteRenter(renterID string) error {
	err := db.deleteInCollectionByID("renters", renterID)
	if err != nil {
		return err
	}
	return nil
}

// Provider operations
//====================

// Return a list of all the providers in the database.
func (db *mongoDB) FindAllProviders() ([]core.ProviderInfo, error) {
	var result []core.ProviderInfo
	err := db.findAllFromCollection("providers", &result)
	if err != nil {
		return nil, err
	}
	if result == nil {
		result = make([]core.ProviderInfo, 0)
	}
	return result, nil
}

// Return the provider with the specified ID.
func (db *mongoDB) FindProviderByID(providerID string) (*core.ProviderInfo, error) {
	var result core.ProviderInfo
	err := db.findOneFromCollectionByID("providers", providerID, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Insert the given provider into the database.
func (db *mongoDB) InsertProvider(provider *core.ProviderInfo) error {
	err := db.insertIntoCollection("providers", provider)
	if err != nil {
		return err
	}
	return nil
}

// Update the given provider in the databse.
func (db *mongoDB) UpdateProvider(provider *core.ProviderInfo) error {
	err := db.updateInCollection("providers", provider.ID, provider)
	if err != nil {
		return err
	}
	return nil
}

// Delete the specified provider from the dtabase.
func (db *mongoDB) DeleteProvider(providerID string) error {
	err := db.deleteInCollectionByID("providers", providerID)
	if err != nil {
		return err
	}
	return nil
}

// File operations
//====================

// Return a list of all files in the database.
func (db *mongoDB) FindAllFiles() ([]core.File, error) {
	var result []core.File
	err := db.findAllFromCollection("files", &result)
	if err != nil {
		return nil, err
	}
	if result == nil {
		result = make([]core.File, 0)
	}
	return result, nil
}

// Return the latest version of the file with the specified ID.
func (db *mongoDB) FindFileByID(fileID string) (*core.File, error) {
	var result core.File
	err := db.findOneFromCollectionByID("files", fileID, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Return a map of paths to files present in the renter's directory.
func (db *mongoDB) FindFilesInRenterDirectory(renterID string) ([]core.File, error) {
	session := db.session.Copy()
	defer session.Close()

	// Get file IDs in renter directory.
	renters := session.DB(dbName).C("renters")

	selector := struct{ ID string }{ID: renterID}
	var renter core.RenterInfo
	err := renters.Find(selector).One(&renter)
	if err != nil {
		return nil, err
	}
	filesToFind := renter.Files

	// Retrieve files from collection.
	files := session.DB(dbName).C("files")
	fileSelector := bson.M{"id": bson.M{"$in": filesToFind}}
	var foundFiles []core.File
	err = files.Find(fileSelector).All(&foundFiles)
	if err != nil {
		return nil, err
	}
	if foundFiles == nil {
		foundFiles = make([]core.File, 0)
	}

	return foundFiles, nil
}

// Return a map of names to files shared with a given renter.
func (db *mongoDB) FindFilesSharedWithRenter(renterID string) ([]core.File, error) {
	session := db.session.Copy()
	defer session.Close()

	// Get file IDs in renter directory.
	renters := session.DB(dbName).C("renters")

	selector := struct{ ID string }{ID: renterID}
	var renter core.RenterInfo
	err := renters.Find(selector).One(&renter)
	if err != nil {
		return nil, err
	}
	filesToFind := renter.Shared

	// Retrieve files from collection.
	files := session.DB(dbName).C("files")
	fileSelector := bson.M{"id": bson.M{"$in": filesToFind}}
	var foundFiles []core.File
	err = files.Find(fileSelector).All(&foundFiles)
	if err != nil {
		return nil, err
	}
	if foundFiles == nil {
		foundFiles = make([]core.File, 0)
	}

	return foundFiles, nil
}

func (db *mongoDB) AddFileToRenterDirectory(renterID string, fileID string) error {
	session := db.session.Copy()
	defer session.Close()

	renters := session.DB(dbName).C("renters")

	selector := bson.M{"id": renterID}
	add := bson.M{"$addToSet": bson.M{"files": fileID}}
	err := renters.Update(selector, add)
	if err != nil {
		return err
	}
	return nil
}

func (db *mongoDB) AddFileToRenterSharedDirectory(renterID string, fileID string) error {
	session := db.session.Copy()
	defer session.Close()

	renters := session.DB(dbName).C("renters")

	selector := bson.M{"id": renterID}
	add := bson.M{"$addToSet": bson.M{"shared": fileID}}
	err := renters.Update(selector, add)
	if err != nil {
		return err
	}
	return nil
}

func (db *mongoDB) RemoveFileFromRenterDirectory(renterID string, fileID string) error {
	session := db.session.Copy()
	defer session.Close()

	renters := session.DB(dbName).C("renters")

	selector := bson.M{"id": renterID}
	pull := bson.M{"$pull": bson.M{"files": fileID}}
	err := renters.Update(selector, pull)
	if err != nil {
		return err
	}
	return nil
}

func (db *mongoDB) RemoveFileFromRenterSharedDirectory(renterID string, fileID string) error {
	session := db.session.Copy()
	defer session.Close()

	renters := session.DB(dbName).C("renters")

	selector := bson.M{"id": renterID}
	pull := bson.M{"$pull": bson.M{"shared": fileID}}
	err := renters.Update(selector, pull)
	if err != nil {
		return err
	}
	return nil
}

// Return a list of files that the renter owns.
func (db *mongoDB) FindFilesByOwner(renterID string) ([]core.File, error) {
	c, session, err := db.getMongoCollection("files")
	if err != nil {
		return nil, err
	}
	defer session.Close()

	var result []core.File
	selector := struct{ ownerID string }{ownerID: renterID}
	err = c.Find(selector).All(&result)
	if err != nil {
		return nil, err
	}
	if result == nil {
		result = make([]core.File, 0)
	}

	return result, nil
}

// Insert the given file into the database.
func (db *mongoDB) InsertFile(file *core.File) error {
	err := db.insertIntoCollection("files", file)
	if err != nil {
		return err
	}
	return nil
}

func (db *mongoDB) InsertFileVersion(fileID string, version *core.Version) error {
	session := db.session.Copy()
	defer session.Close()

	files := session.DB(dbName).C("files")
	currentVersions := session.DB(dbName).C("versions")

	selector := bson.M{"id": fileID}
	// Atomically get the next version number from the currentVersions collection
	versionUpdate := bson.M{"$inc": bson.M{"number": 1}}
	versionChange := mgo.Change{
		Update:    versionUpdate,
		Upsert:    true,
		ReturnNew: true,
	}
	var result fileVersion
	_, err := currentVersions.Find(selector).Apply(versionChange, &result)
	if err != nil {
		return err
	}
	version.Num = result.Number
	push := bson.M{"$push": bson.M{"versions": version}}
	err = files.Update(selector, push)
	if err != nil {
		return err
	}
	return nil
}

// Update the given file in the database.
func (db *mongoDB) UpdateFile(file *core.File) error {
	err := db.updateInCollection("files", file.ID, file)
	if err != nil {
		return err
	}
	return nil
}

func (db *mongoDB) RenameFolder(fileID string, ownerID string, oldName string, newName string) error {
	session := db.session.Copy()
	defer session.Close()

	files := session.DB(dbName).C("files")

	findRegex := fmt.Sprintf("^%s/", oldName)
	selector := bson.M{"name": bson.M{"$regex": findRegex}, "ownerid": ownerID}

	var result []core.File
	err := files.Find(selector).All(&result)
	if err != nil {
		return err
	}

	r, err := regexp.Compile(findRegex)
	if err != nil {
		return err
	}

	newPrefix := fmt.Sprintf("%s/", newName)
	for _, item := range result {
		newString := r.ReplaceAllString(item.Name, newPrefix)
		item.Name = newString
		err = db.UpdateFile(&item)
		if err != nil {
			return err
		}
	}

	return nil
}

func (db *mongoDB) RemoveFolderChildren(folder *core.File) ([]core.File, error) {
	session := db.session.Copy()
	defer session.Close()

	files := session.DB(dbName).C("files")

	findRegex := fmt.Sprintf("^%s/", folder.Name)
	selector := bson.M{"name": bson.M{"$regex": findRegex}, "ownerid": folder.OwnerID}

	var removed []core.File
	err := files.Find(selector).All(&removed)
	if err != nil {
		return nil, err
	}
	if removed == nil {
		removed = make([]core.File, 0)
	}

	_, err = files.RemoveAll(selector)
	if err != nil {
		return nil, err
	}

	return removed, nil
}

func (db *mongoDB) UpdateFileVersion(fileID string, version *core.Version) error {
	session := db.session.Copy()
	defer session.Close()

	files := session.DB(dbName).C("files")

	selector := bson.M{"id": fileID, "versions.num": version.Num}
	// Atomically get the next version number from the currentVersions collection
	versionUpdate := bson.M{"$set": bson.M{"versions.$": version}}
	err := files.Update(selector, versionUpdate)
	if err != nil {
		return err
	}
	return nil
}

// Delete all versions of the given file from the database.
func (db *mongoDB) DeleteFile(fileID string) error {
	err := db.deleteInCollectionByID("files", fileID)
	if err != nil {
		return err
	}
	return nil
}

func (db *mongoDB) DeleteFileVersion(fileID string, version int) error {
	session := db.session.Copy()
	defer session.Close()

	files := session.DB(dbName).C("files")

	selector := bson.M{"id": fileID}
	pull := bson.M{"$pull": bson.M{"versions": bson.M{"num": version}}}
	err := files.Update(selector, pull)
	if err != nil {
		return err
	}
	return nil
}

func (db *mongoDB) AddPermissionToFileACL(fileID string, permission *core.Permission) error {
	session := db.session.Copy()
	defer session.Close()

	files := session.DB(dbName).C("files")

	selector := bson.M{"id": fileID, "accesslist.renterid": bson.M{"$ne": permission.RenterId}}
	push := bson.M{"$push": bson.M{"accesslist": permission}}
	err := files.Update(selector, push)
	if err != nil {
		return err
	}
	return nil
}

func (db *mongoDB) UpdateFilePermission(fileID string, permission *core.Permission) error {
	session := db.session.Copy()
	defer session.Close()

	files := session.DB(dbName).C("files")

	selector := bson.M{"id": fileID, "accesslist.renterid": permission.RenterId}
	// Atomically get the next version number from the currentVersions collection
	update := bson.M{"$set": bson.M{"accesslist.$": permission}}
	err := files.Update(selector, update)
	if err != nil {
		return err
	}
	return nil
}

func (db *mongoDB) RemoveFilePermissionFromACL(fileID string, renterID string) error {
	session := db.session.Copy()
	defer session.Close()

	files := session.DB(dbName).C("files")

	selector := bson.M{"id": fileID, "accesslist.renterid": renterID}
	pull := bson.M{"$pull": bson.M{"accesslist": bson.M{"renterid": renterID}}}
	err := files.Update(selector, pull)
	if err != nil {
		return err
	}
	return nil
}

// Contract operations
//====================

// Return a list of all contracts in the database.
func (db *mongoDB) FindAllContracts() ([]core.Contract, error) {
	var result []core.Contract
	err := db.findAllFromCollection("contracts", &result)
	if err != nil {
		return nil, err
	}
	if result == nil {
		result = make([]core.Contract, 0)
	}
	return result, nil
}

// Return the contract with the specified ID
func (db *mongoDB) FindContractByID(contractID string) (*core.Contract, error) {
	var result core.Contract
	err := db.findOneFromCollectionByID("contracts", contractID, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Return a list of contracts belonging to the specified renter.
func (db *mongoDB) FindContractsByRenter(renterID string) ([]core.Contract, error) {
	c, session, err := db.getMongoCollection("contracts")
	if err != nil {
		return nil, err
	}
	defer session.Close()

	var result []core.Contract
	selector := bson.M{"renterid": renterID}
	err = c.Find(selector).All(&result)
	if err != nil {
		return nil, err
	}
	if result == nil {
		result = make([]core.Contract, 0)
	}

	return result, nil
}

// Insert the given contract into the database.
func (db *mongoDB) InsertContract(contract *core.Contract) error {
	err := db.insertIntoCollection("contracts", contract)
	if err != nil {
		return err
	}
	return nil
}

// Update the given contract.
func (db *mongoDB) UpdateContract(contract *core.Contract) error {
	err := db.updateInCollection("contracts", contract.ID, contract)
	if err != nil {
		return err
	}
	return nil
}

// Delete the contract.
func (db *mongoDB) DeleteContract(contractID string) error {
	err := db.deleteInCollectionByID("contracts", contractID)
	if err != nil {
		return err
	}
	return nil
}

// Payment operations
//====================

// Return a list of all payments in the database.
func (db *mongoDB) FindAllPayments() ([]core.PaymentInfo, error) {
	var result []core.PaymentInfo
	err := db.findAllFromCollection("payments", &result)
	if err != nil {
		return nil, err
	}
	if result == nil {
		result = make([]core.PaymentInfo, 0)
	}
	return result, nil
}

// Find the payment associated with the given contract.
func (db *mongoDB) FindPaymentByContract(contractID string) (*core.PaymentInfo, error) {
	session := db.session.Copy()
	defer session.Close()

	c := session.DB(dbName).C("payments")

	selector := bson.M{"contract": contractID}
	var result core.PaymentInfo
	err := c.Find(selector).One(&result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// Insert the given payment into the database.
func (db *mongoDB) InsertPayment(payment *core.PaymentInfo) error {
	err := db.insertIntoCollection("payments", payment)
	if err != nil {
		return err
	}
	return nil
}

// Update the given payment.
func (db *mongoDB) UpdatePayment(payment *core.PaymentInfo) error {
	session := db.session.Copy()
	defer session.Close()

	c := session.DB(dbName).C("payments")
	selector := bson.M{"contractid": payment.ContractID}
	err := c.Update(selector, payment)
	if err != nil {
		return err
	}
	return nil
}

// Delete the payment.
func (db *mongoDB) DeletePayment(contractID string) error {
	session := db.session.Copy()
	defer session.Close()

	c := session.DB(dbName).C("payments")
	selector := bson.M{"contract": contractID}
	err := c.Remove(selector)
	if err != nil {
		return err
	}
	return nil
}

// Transaction operations
//=======================

// Return a list of all transactions in the database.
func (db *mongoDB) FindAllTransactions() ([]core.Transaction, error) {
	result := make([]core.Transaction, 0)
	err := db.findAllFromCollection("transactions", &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Find the transactions associated with given renter.
func (db *mongoDB) FindTransactionsByRenter(renterID string) ([]core.Transaction, error) {
	session := db.session.Copy()
	defer session.Close()

	c := session.DB(dbName).C("transactions")

	selector := bson.M{"usertype": "renter", "userid": renterID}
	result := make([]core.Transaction, 0)
	err := c.Find(selector).All(&result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// Find the transactions associated with given provider.
func (db *mongoDB) FindTransactionsByProvider(providerID string) ([]core.Transaction, error) {
	session := db.session.Copy()
	defer session.Close()

	c := session.DB(dbName).C("transactions")

	selector := bson.M{"usertype": "provider", "userid": providerID}
	result := make([]core.Transaction, 0)
	err := c.Find(selector).All(&result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// Insert the given transaction into the database.
func (db *mongoDB) InsertTransaction(transaction *core.Transaction) error {
	err := db.insertIntoCollection("transactions", transaction)
	if err != nil {
		return err
	}
	return nil
}
