## Usage 

```
go build [-o output]
# Run new executable
```

## Execution time

The extraction of all available data took approximately 6 hours. The data weighted 38.2 MB.

## Scope

Edit the file  ```ubigeos.json``` to limit the scope of the data gathered to include only the locations you are interested in.

## About the fetching

The program spawns a goroutine for each ubigeo to compute the data concurrently.


## Resulting files

**\* These files only include products classified as COVID-19 essential items.**

### list.json

They key is composite using the product code ('codprod') and drugstore code ('codigo').

```json
{
	"10893-0023149": {
		"codigo": "0023149",
		"codprod": 10893,
		"farmacia": "MIFARMA",
		"fecha": "19/08/2020 13:16:05",
		"laboratorio": "MEDIFARMA",
		"marca": "",
		"modelo": "",
		"nombre": "TRI AZIT 500 mg Tableta",
		"pre_max": null,
		"pre_med": null,
		"pre_min": null,
		"precio": 1.3,
		"regsan": "N17940",
		"segmento": "0",
		"setcodigo": "PRIVADO",
		"suscomun": "AZITROMICINA"
    }
}
```

### drugstores.json

They key is the drugstore code and 'codigo' key in each list.json entry.

```json
{
  "0000275": {
    "codest": "0012732",
    "direccion": "AV. LA ENCALADA CENTRO COMERCIAL MONTERRICO 640 MZ. R - 1, LT. 10",
    "horario": "LUN A VIE: 07:00 A 23:00; SAB: 07:00 A 23:00; DOM: 07:00 A 23:00",
    "nombre": "BOTICA BOTICAS FASA",
    "ruc": "20512002090",
    "telefono": "4117777",
    "tipo": "M",
    "ubicacion": "LIMA - SANTIAGO DE SURCO"
  }
}
```

### products.json

They key is the product code and 'codprod' key in each list.json entry.

```json
{
  "10983": {
    "a": "PARACETAMOL",
    "b": "DOLOACEMIFEN",
    "c": "500 mg",
    "d": "TABLETA",
    "e": "TABLETA",
    "f": "Caja Envase Blister Tabletas",
    "g": "1 TABLETA",
    "h": "EN00378",
    "i": "",
    "j": "SHERFARMA",
    "k": "LABORATORIOS PORTUGAL S.R.L.",
    "l": "PARACETAMOL 500 MG TABLETA"
  }
}
```

## Key references 

```
	Ruc             -> `json:"ruc"`
	Name            -> `json:"nombre"`
	Address         -> `json:"direccion"`
	Location        -> `json:"ubicacion"`
	Type            -> `json:"tipo"`
	Phone           -> `json:"telefono"`
	OpenHours       -> `json:"horario"`
	DrugStoreId     -> `json:"codigo"`
	GenericName     -> `json:"nombre"`
	MarketName      -> `json:"b"`
	Concentration   -> `json:"c"`
	Form            -> `json:"d"`
	Presentations   -> `json:"f"`
	Laboratory      -> `json:"laboratorio"`
	Manufacturer    -> `json:"k"`
	SearchName      -> `json:"l"`
	UpdatedAt       -> `json:"fecha"`
	Sector          -> `json:"setcodigo"`
	ProductId       -> `json:"codprod"`
	HealthRegistry  -> `json:"regsan"`
```

