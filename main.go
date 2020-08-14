package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

const ProductsUrl = "https://opmcovid.minsa.gob.pe/observatorio/precios.aspx/GetMedicine"
const PricesUrl = "https://opmcovid.minsa.gob.pe/observatorio/wsObservatorio.asmx/listPrice"
const PharmaUrl = "https://opmcovid.minsa.gob.pe/observatorio/wsObservatorio.asmx/loadDataPharma"

type productPrice struct {
	Product product `json:"product"`
	Pharma  pharma  `json:"pharma"`
}

type pharma struct {
	Ruc       string `json:"ruc"`
	Name      string `json:"nombre"`
	Address   string `json:"direccion"`
	Location  string `json:"ubicacion"`
	Type      string `json:"tipo"`
	Phone     string `json:"telefono"`
	OpenHours string `json:"horario"`
}

type productList struct {
	Data []product `json:"products"`
}

type product struct {
	PharmaId       string `json:"codigo"`
	GenericName    string `json:"nombre"`
	MarketName     string `json:"b"`
	Concentration  string `json:"c"`
	Form           string `json:"d"`
	Presentations  string `json:"f"`
	Laboratory     string `json:"laboratorio"`
	Manufacturer   string `json:"k"`
	SearchName     string `json:"l"`
	UpdatedAt      string `json:"fecha"`
	Sector         string `json:"setcodigo"`
	ProductId      int    `json:"codprod"`
	HealthRegistry string `json:"regsan"`
}

var productMap = make(map[int]product)
var pharmaMap = make(map[string]pharma)

func main() {
	names := getProductsName()
	name := strings.Split(names[4], "-")[0]
	list := getPriceList(name, "01", 1)

	dataBytes, err := json.Marshal(list)
	if err != nil {
		panic(err)
	}

	err = ioutil.WriteFile("one_big_fucking_product.json", dataBytes, 0644)
	if err != nil {
		panic(err)
	}
}

func fetchWrapper(url, body string) []string {
	search := []byte(body)
	r, err := http.Post(url, "application/json", bytes.NewBuffer(search))
	if err != nil {
		panic(err)
	}

	obj := make(map[string][]string)
	err = json.NewDecoder(r.Body).Decode(&obj)

	return obj["d"]
}

func getPriceList(name, ubigeo string, fromPage int) []productPrice {
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

	var productList productList
	data := []byte(fmt.Sprintf(`{"products": %s}`, resp[0]))
	if err := json.Unmarshal(data, &productList); err != nil {
		panic(err)
	}

	var productPrices []productPrice

	for _, product := range productList.Data {
		var productPrice productPrice

		pharm, p := getPharma(product.PharmaId, product.ProductId)

		productPrice.Pharma = pharm

		product.MarketName = p.MarketName
		product.Concentration = p.Concentration
		product.Form = p.Form
		product.Presentations = p.Presentations
		product.Manufacturer = p.Manufacturer
		product.SearchName = p.SearchName

		productPrice.Product = product

		productPrices = append(productPrices, productPrice)
	}

	re := regexp.MustCompile(`\d+`)
	currentPage, _ := strconv.Atoi(string(re.Find([]byte(resp[1]))))

	if fromPage != currentPage {
		productPrices = append(productPrices, getPriceList(name, ubigeo, fromPage+1)...)
	}

	return productPrices
}

func getPharma(pharmaId string, productId int) (pharma, product) {

	pharm, hasPharm := pharmaMap[pharmaId]
	product, hasProduct := productMap[productId]

	if hasPharm && hasProduct {
		return pharm, product
	}

	search := fmt.Sprintf(`{"cod_estab":"%s","cod_prod":%d}`, pharmaId, productId)
	resp := fetchWrapper(PharmaUrl, search)
	pharmaData := resp[0][1 : len(resp[0])-1]
	productData := resp[1][1 : len(resp[1])-1]

	err := json.Unmarshal([]byte(pharmaData), &pharm)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal([]byte(productData), &product)
	if err != nil {
		panic(err)
	}

	pharmaMap[pharmaId] = pharm
	productMap[productId] = product

	return pharm, product
}

func getProductsName() []string {
	search := "{'prefix': ''}"
	return fetchWrapper(ProductsUrl, search)
}
