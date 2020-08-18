package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/schollz/progressbar/v3"
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"
)

const ProductsUrl = "https://opmcovid.minsa.gob.pe/observatorio/precios.aspx/GetMedicine"
const PricesUrl = "https://opmcovid.minsa.gob.pe/observatorio/wsObservatorio.asmx/listPrice"
const PharmaUrl = "https://opmcovid.minsa.gob.pe/observatorio/wsObservatorio.asmx/loadDataPharma"

type productPrice struct {
	Product   Product   `json:"Product"`
	DrugStore DrugStore `json:"drugstore"`
}

type ubigeo struct {
	Id   int            `json:"id_ubigeo"`
	Data []productPrice `json:"data"`
}

type DrugStore struct {
	Ruc       string `json:"ruc"`
	Name      string `json:"nombre"`
	Address   string `json:"direccion"`
	Location  string `json:"ubicacion"`
	Type      string `json:"tipo"`
	Phone     string `json:"telefono"`
	OpenHours string `json:"horario"`
}

type Product struct {
	DrugStoreId    string  `json:"codigo"`
	GenericName    string  `json:"nombre"`
	MarketName     string  `json:"b"`
	Concentration  string  `json:"c"`
	Price          float64 `json:"precio"`
	Form           string  `json:"d"`
	Presentations  string  `json:"f"`
	Laboratory     string  `json:"laboratorio"`
	Manufacturer   string  `json:"k"`
	SearchName     string  `json:"l"`
	UpdatedAt      string  `json:"fecha"`
	Sector         string  `json:"setcodigo"`
	ProductId      int     `json:"codprod"`
	HealthRegistry string  `json:"regsan"`
}

// var wg sync.WaitGroup
var client = retryablehttp.NewClient()

func main() {
	client.Logger = nil
	client.RetryMax = 10
	ubigeos := getAllUbigeos()
	names := getProductsName()
	productsUbigeo := make(chan ubigeo, len(ubigeos))
	bar := progressbar.Default(int64(len(ubigeos) * len(names)))
	done := 0

	for _, ubigeo := range ubigeos {
		go generatePriceListByUbigeo(strconv.Itoa(ubigeo.Id), names, productsUbigeo, &done)
	}

	for {
		_ = bar.Set(done)
		if len(productsUbigeo) == cap(productsUbigeo) {
			// Channels are full
			break
		}
	}

	close(productsUbigeo)

	var result []ubigeo

	for u := range productsUbigeo {
		result = append(result, u)
	}

	dataBytes, err := json.Marshal(result)
	if err != nil {
		panic(err)
	}

	err = ioutil.WriteFile("one_big_fucking_product.json", dataBytes, 0644)
	if err != nil {
		panic(err)
	}

}

func generatePriceListByUbigeo(ubigeoId string, productsNames []string, c chan ubigeo, done *int) {
	//defer wg.Done()
	var drugStoreMap = make(map[string]DrugStore)
	var productsMap = make(map[int]Product)
	ubiID, _ := strconv.Atoi(ubigeoId)
	result := ubigeo{Id: ubiID}
	for _, name := range productsNames {
		products := getProductPrices(name, ubigeoId, 1)
		for _, p := range products {
			drugstore, product := fillProductData(p, productsMap, drugStoreMap)
			finalList := productPrice{DrugStore: drugstore, Product: product}
			result.Data = append(result.Data, finalList)
			*done += 1
		}
	}
	c <- result
}

func fetchWrapper(url, body string) []string {
	search := []byte(body)

	r, err := client.HTTPClient.Post(url, "application/json", bytes.NewBuffer(search))
	if err != nil {
		panic(err)
	}
	defer r.Body.Close()

	obj := make(map[string][]string)
	err = json.NewDecoder(r.Body).Decode(&obj)

	return obj["d"]
}

func getProductPrices(name, ubigeo string, fromPage int) []Product {
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
`, name, ubigeo, fromPage)

	resp := fetchWrapper(PricesUrl, search)

	type productList struct {
		Data []Product `json:"products"`
	}

	var products productList

	data := []byte(fmt.Sprintf(`{"products": %s}`, resp[0]))
	if err := json.Unmarshal(data, &products); err != nil {
		panic(err)
	}

	re := regexp.MustCompile(`\d+`)
	currentPage, _ := strconv.Atoi(string(re.Find([]byte(resp[1]))))

	if fromPage < currentPage {
		products.Data = append(products.Data, getProductPrices(name, ubigeo, fromPage+1)...)
	}

	return products.Data
}

func fillProductData(firstHalf Product, productMap map[int]Product, drugStoreMap map[string]DrugStore) (DrugStore, Product) {
	drugStore, hasDrugStore := drugStoreMap[firstHalf.DrugStoreId]
	product, hasProduct := productMap[firstHalf.ProductId]

	if hasDrugStore && hasProduct {
		return drugStore, product
	}

	drugStore, secondHalf := getDrugStore(firstHalf.DrugStoreId, firstHalf.ProductId)
	drugStoreMap[firstHalf.DrugStoreId] = drugStore

	if secondHalf != (Product{}) {
		firstHalf.MarketName = secondHalf.MarketName
		firstHalf.Concentration = secondHalf.Concentration
		firstHalf.Form = secondHalf.Form
		firstHalf.Presentations = secondHalf.Presentations
		firstHalf.Manufacturer = secondHalf.Manufacturer
		firstHalf.SearchName = secondHalf.SearchName

		productMap[firstHalf.ProductId] = firstHalf
	}

	return drugStore, firstHalf
}

func getDrugStore(drugStoreId string, productId int) (DrugStore, Product) {

	search := fmt.Sprintf(`{"cod_estab":"%s","cod_prod":%d}`, drugStoreId, productId)
	resp := fetchWrapper(PharmaUrl, search)
	drugStoreData := resp[0][1 : len(resp[0])-1]
	productData := fmt.Sprintf(`{"products": %s}`, resp[1])

	type productWrapper struct {
		Products []Product `json:"products"`
	}

	var drugStore DrugStore
	err := json.Unmarshal([]byte(drugStoreData), &drugStore)
	if err != nil {
		panic(err)
	}

	var pw productWrapper
	err = json.Unmarshal([]byte(productData), &pw)
	if err != nil {
		panic(err)
	}

	if len(pw.Products) == 0 {
		pw.Products = append(pw.Products, Product{})
	}

	return drugStore, pw.Products[0]
}

func getProductsName() []string {
	search := "{'prefix': ''}"
	raw := fetchWrapper(ProductsUrl, search)
	var names []string

	for _, fullName := range raw {

		short := strings.ReplaceAll(strings.Split(fullName, "-")[0], "\"", "")
		if len(short) < 5 {
			continue
		}

		names = append(names, short)
	}

	return names
}

func getAllUbigeos() []ubigeo {
	file, _ := ioutil.ReadFile("ubigeos.json")

	var ubigeos []ubigeo
	err := json.Unmarshal(file, &ubigeos)
	if err != nil {
		panic(err)
	}

	return ubigeos
}
