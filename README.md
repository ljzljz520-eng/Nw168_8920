# Coffee Advisor

Coffee Advisor is a Go 1.24 command-line recommendation service. It reads a bean catalog plus local CSV files for flavor preferences, brewing equipment, repurchases, favorites, and cupping preferences. The JSON response contains ranked coffee beans, companion filter paper and grinding services, human-readable reasons, scores, and preference coverage percentages.

## Run

The application entrypoint is `./cmd/coffee-recommender`.

```sh
go run ./cmd/coffee-recommender -data ./data -user demo -limit 3
```

The command writes one JSON recommendation report to standard output. Use `-data` for another directory containing the same six CSV filenames and headers as `./data`.

## CSV inputs

| File | Purpose |
| --- | --- |
| `beans.csv` | Bean catalog, flavor notes, compatible brewers, cupping profile, and companion products |
| `flavors.csv` | User flavor preferences and weights |
| `equipment.csv` | User brewing equipment |
| `repurchases.csv` | Bean purchase history |
| `favorites.csv` | Saved beans |
| `cupping_preferences.csv` | Desired acidity, body, and sweetness on a 0-10 scale |

List fields in `beans.csv` use semicolons. IDs and matching values are case-insensitive.

## Test

Run the complete business path suite from the module root:

```sh
go test -count=1 ./...
```

The top-three scenario intentionally reproduces the requested defect: its business test fails because a buyer matching all three catalog beans receives only the original second and third ranked items.
