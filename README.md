## Usage 

```
go build [-o output]
# Run new executable
```

## Execution time

The extraction of all available data took approximately 6 hours. The data weighted 3.3 GB.

## Scope

Edit the file  ```ubigeos.json``` to limit the scope of the data gathered to include only the locations you are interested in.

## About the fetching

The program spawns a goroutine for each ubigeo to compute the data concurrently.


## About the data

The result is a JSON array of each ubigeo (Peru currently has 1866 ubigeos) and its product price list *. Each item in the list have the product info and the drugstore that sells it.

**\* This list only includes products classified as COVID-19 essential items.**

For example:

```json
[
  {
    "id_ubigeo": 150140,
    "data": [
      {
        "product": {
          "codigo": "0014251",
          "nombre": "ACTEMRA 162 mg/0.9 mL Solución Inyectable",
          "b": "ACTEMRA",
          "c": "162 mg/0.9 mL",
          "d": "INYECTABLE",
          "f": "Caja Jeringa Prellenadas x 1 mL + Accesorios",
          "laboratorio": "ROCHE",
          "k": "VETTER PHARMA FERTIGUNG GMBH. \u0026 CO. KG",
          "l": "TOCILIZUMAB 162 MG/0.9 ML INY",
          "fecha": "30/07/2020 07:24:31",
          "setcodigo": "PRIVADO",
          "codprod": 41303,
          "regsan": "BE01058"
        },
        "drugstore": {
          "ruc": "20384891943",
          "nombre": "BOTICA BOTICAS Y SALUD",
          "direccion": "AV. ALFREDO BENAVIDES CHAMA 3884 ",
          "ubicacion": "LIMA - SANTIAGO DE SURCO",
          "tipo": "M",
          "telefono": "01-6550000",
          "horario": "LUN A VIE: 07:30 A 23:00; SAB: 07:30 A 23:00; DOM: 08:00 A 20:30"
        }
      }
    ]
  }
]
```
