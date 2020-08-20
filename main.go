package main

import (
	"bytes"
	"fmt"
	"github.com/Jeffail/gabs/v2"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/schollz/progressbar/v3"
	"io/ioutil"
	"reflect"
	"strconv"
	"strings"
	"sync/atomic"
)

const ProductsUrl = "https://opmcovid.minsa.gob.pe/observatorio/precios.aspx/GetMedicine"
const PricesUrl = "https://opmcovid.minsa.gob.pe/observatorio/wsObservatorio.asmx/listPrice"
const PharmaUrl = "https://opmcovid.minsa.gob.pe/observatorio/wsObservatorio.asmx/loadDataPharma"

type Result struct {
	ubigeoId   int
	products   map[int]*gabs.Container
	drugstores map[string]*gabs.Container
	resultList map[string]*gabs.Container
}

var client = retryablehttp.NewClient()

func main() {
	client.Logger = nil
	client.RetryMax = 10

	ubigeos := readUbigeos()
	names := getProductsName()

	listByUbigeo := make(chan Result, len(ubigeos))

	bar := progressbar.Default(int64(len(ubigeos) * len(names)))
	var done uint64

	for _, ubigeoId := range ubigeos {
		go generateListByUbigeo(ubigeoId, names, listByUbigeo, &done)
	}

	for {
		_ = bar.Set(int(done))
		if len(listByUbigeo) == cap(listByUbigeo) {
			// Channels are full
			break
		}
	}

	close(listByUbigeo)

	resultObj := gabs.New()
	productsObj := gabs.New()
	drugstoresObj := gabs.New()

	for u := range listByUbigeo {

		_, err := resultObj.Set(u.resultList, strconv.Itoa(u.ubigeoId))
		if err != nil {
			panic(err)
		}

		_, err = productsObj.Set(u.products)
		if err != nil {
			panic(err)
		}

		_, err = drugstoresObj.Set(u.drugstores)
		if err != nil {
			panic(err)
		}

	}

	writeFile("products.json", productsObj.Bytes())
	writeFile("drugstores.json", drugstoresObj.Bytes())
	writeFile("list.json", resultObj.Bytes())
}

func writeFile(name string, content []byte) {
	err := ioutil.WriteFile(name, content, 0644)
	if err != nil {
		panic(err)
	}
}

func generateListByUbigeo(ubigeoId int, productsNames []string, c chan Result, done *uint64) {
	var drugStoreMap = make(map[string]*gabs.Container)
	var productsMap = make(map[int]*gabs.Container)
	var list = make(map[string]*gabs.Container)

	for _, name := range productsNames {
		products := getList(name, ubigeoId, 1)
		for _, item := range products {
			productCode := int(item.S("codprod").Data().(float64))
			product, hasProduct := productsMap[productCode]

			drugstoreCode := item.S("codigo").Data().(string)
			drugstore, hasDrugstore := drugStoreMap[drugstoreCode]

			if !(hasProduct && hasDrugstore) {
				product, drugstore = getDrugstore(drugstoreCode, productCode)

				if !reflect.DeepEqual(product, gabs.New()) {
					productsMap[productCode] = product
				}

				if !reflect.DeepEqual(drugstore, gabs.New()) {
					drugStoreMap[drugstoreCode] = drugstore
				}
			}

			combinedKey := fmt.Sprintf("%d-%s", productCode, drugstoreCode)
			list[combinedKey] = item
		}
		atomic.AddUint64(done, 1)
	}

	result := Result{ubigeoId: ubigeoId, products: productsMap, drugstores: drugStoreMap, resultList: list}

	c <- result
}

func getList(name string, ubigeo int, fromPage int) []*gabs.Container {
	search := fmt.Sprintf(`
	{
		"typeup": "FARMACIA",
		"tipo": "M",
		"nomprod": "%s",
		"ubigeo": "%s",
		"type": "0",
		"labotarorio": "",
		"establecimiento": "",
		"pag": %d
	}
	`, name, strconv.Itoa(ubigeo), fromPage)

	resp := fetchWrapper(PricesUrl, search).Children()

	list, err := gabs.ParseJSON([]byte(resp[0].Data().(string)))
	if err != nil {
		panic(err)
	}

	result := list.Children()

	pages, err := gabs.ParseJSON([]byte(resp[1].Data().(string)))
	if err != nil {
		panic(err)
	}

	lastPage := int(pages.Children()[0].S("tpaginas").Data().(float64))

	if fromPage < lastPage {
		result = append(result, getList(name, ubigeo, fromPage+1)...)
	}

	return result
}

func getDrugstore(drugStoreId string, productId int) (drugstore, product *gabs.Container) {
	search := fmt.Sprintf(`{"cod_estab":"%s","cod_prod":%d}`, drugStoreId, productId)
	resp := fetchWrapper(PharmaUrl, search).Children()

	rawDrugstore, err := gabs.ParseJSON([]byte(resp[0].Data().(string)))
	if err != nil {
		panic(err)
	}

	if len(rawDrugstore.Children()) == 0 {
		drugstore = gabs.New()
	} else {
		drugstore = rawDrugstore.Children()[0]
	}

	rawProduct, err := gabs.ParseJSON([]byte(resp[1].Data().(string)))
	if err != nil {
		panic(err)
	}

	if len(rawProduct.Children()) == 0 {
		product = gabs.New()
	} else {
		product = rawProduct.Children()[0]
	}

	return
}

func getProductsName() []string {
	search := "{'prefix': ''}"
	raw := fetchWrapper(ProductsUrl, search)
	namesSet := make(map[string]bool)

	for _, fullName := range raw.Children() {
		short := strings.ReplaceAll(strings.Split(fullName.Data().(string), "-")[0], "\"", "")
		if len(short) < 5 {
			continue
		}

		_, ok := namesSet[short]
		if !ok {
			namesSet[short] = true
		}
	}

	var names []string

	for name := range namesSet {
		names = append(names, name)
	}

	return names
}

func readUbigeos() []int {
	file, _ := ioutil.ReadFile("ubigeos.json")

	ubigeosJson, err := gabs.ParseJSON(file)
	if err != nil {
		panic(err)
	}

	var ubigeos []int

	for _, ubigeo := range ubigeosJson.Children() {
		ubigeos = append(ubigeos, int(ubigeo.S("id_ubigeo").Data().(float64)))
	}

	return ubigeos
}

func fetchWrapper(url, search string) *gabs.Container {
	resp, err := client.HTTPClient.Post(url, "application/json", bytes.NewBuffer([]byte(search)))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	raw, err := gabs.ParseJSON(body)
	if err != nil {
		panic(err)
	}

	return raw.Path("d")
}
