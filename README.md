# Nuolydis

Eksperimentinė azartinių lošimų mažinimo priemonė. Šioje repozitorijoje yra II užduoties veikiantis prototipas: taisyklių ir trinties variklis, ruletė virtualiais taškais, duomenų atskyrimas tarp klientų ir paprasta naršyklės sąsaja.

Tai nėra gydymas. Prototipas naudoja tik sintetinius duomenis. Projektavimo dokumentas: [I-uzduotis.md](I-uzduotis.md).

## Reikalingi įrankiai

Veikia „Windows“ ir „macOS“.

| Įrankis | Versija | Įdiegimas |
| --- | --- | --- |
| Go | 1.26 arba naujesnė | <https://go.dev/dl/> |
| Node.js | 26 LTS | `winget install OpenJS.NodeJS.LTS` arba `brew install node@26` |
| Docker Desktop | dabartinė, turi būti paleista | <https://www.docker.com/products/docker-desktop/> |
| golangci-lint | v2 (tik kodo tikrinimui) | <https://golangci-lint.run/welcome/install/> |

## Paleidimas

```
go run ./cmd/dev
```

Komanda paleidžia MySQL, sukompiliuoja sąsają ir paleidžia serverį. Atidarykite <http://localhost:8080>.

Pavyzdinių duomenų atkūrimas (būtina prieš kiekvieną demonstraciją):

```
go run ./cmd/dev -reset
```

## Konfigūracija

Visi nustatymai neprivalomi, numatytosios reikšmės atitinka `docker-compose.yml`. Pavyzdys yra [.env.example](.env.example). API raktų nereikia.

| Kintamasis | Numatytoji reikšmė | Paskirtis |
| --- | --- | --- |
| `HTTP_ADDR` | `:8080` | Serverio adresas |
| `MYSQL_DSN` | žr. `.env.example` | Duomenų bazės prisijungimas (prievadas 3307) |
| `DEMO_MODE` | `true` | Įjungia savaitės keitimo demonstracijos maršrutą |

## Testai

```
go test ./...                                  # vienetų testai, „Docker“ nereikia
docker compose up -d --wait                    # integraciniams testams
go test -tags=integration -p 1 ./...           # įskaitant klientų izoliavimo testus
go vet ./...
golangci-lint run
```

Sąsajos patikros, aplanke `web`: `npm run lint`, `npm run typecheck`, `npm run build`.

Plačiau: [docs/testing-plan.md](docs/testing-plan.md).

## Pavyzdiniai duomenys

Įrašomi automatiškai, kai duomenų bazė tuščia.

| Klientas | Būsena | Skirta parodyti |
| --- | --- | --- |
| Pora A | 1 plano savaitė, skyrimas ir balansas 500, didžiausias statymas 50, nuostolio riba 200 | Įprastas statymas, plano mažinimas |
| Pora B | 9 plano savaitė, skyrimas 290, balansas 250, atvira sesija su 2 vijimosi signalais ir 2 įspėjimais, paskutinis statymas 20 pralaimėtas | Vijimosi atmetimas ir užrakinimas iki pirmadienio |
| Pora C | 9 plano savaitė, skyrimas 290, balansas 250, be sesijų | Aukščiausias trinties lygis |

## Demonstravimo žingsniai

1. Paleiskite `go run ./cmd/dev -reset` ir atidarykite <http://localhost:8080>.
2. **Pagrindinis scenarijus.** Pora A, „Pradėti sesiją“, statymas 20, pasirinkimas `red`, „Statyti“. Rezultatas `ACCEPTED`, balansas pasikeičia pagal raundą.
3. **Plano mažinimas.** „Baigti sesiją“, tada „Demonstracija: pereiti į kitą savaitę“. Skyrimas 420, pažanga 16 %. Tą patį matote prižiūrėtojo veiksenoje.
4. **Konfliktas.** Pora B, statymas 40, pasirinkimas `red`, „Statyti“. Rezultatas `REJECTED`, `CHASING_LIMIT`, sesija baigta. Prižiūrėtojo veiksenoje matomas įspėjimas. Bandant pradėti naują sesiją: `DENIED`, `CHASING_LOCKOUT`, rodoma kito skyrimo data.
5. **Klaida.** Pora C, nauja sesija, statymas 20. Prašoma įvesti sakinį. Įveskite neteisingą: `CONFIRMATION_FAILED`, balansas nepakitęs. Įveskite teisingą: statymas priimamas.
6. **Klientų atskyrimas.** Pirmiau Poros A kontekste skaitomi Poros B duomenys. Tenant ID sužinomi iš `GET /api/tenants`, Pora A yra 1, Pora B yra 2:

   ```
   curl -i -H "X-Tenant-ID: 1" http://localhost:8080/api/tenants/2/state
   ```

   Atsakymas `403`. Poros B sesijos ID per Poros A kelią:

   ```
   curl -i -X POST -H "X-Tenant-ID: 1" http://localhost:8080/api/tenants/1/sessions/1/end
   ```

   Atsakymas `404`, duomenys nepakeisti.
7. **Testai.** `go test -tags=integration -p 1 ./...`

## Struktūra

| Aplankas | Paskirtis |
| --- | --- |
| `internal/model` | Duomenų struktūros, sprendimai ir priežasčių kodai |
| `internal/service/engine` | Pagrindinis modulis: taisyklių konvejeris, be išorinių priklausomybių |
| `internal/service/game` | `Game` sąsaja ir ruletė |
| `internal/service` | Scenarijų koordinavimas ir sąsajos, kurias realizuoja infrastruktūra |
| `internal/infra` | MySQL (GORM), laikrodis, atsitiktinių skaičių šaltinis |
| `internal/api` | HTTP valdikliai, kliento konteksto tikrinimas, sąsajos paslaugoms |
| `openapi` | API aprašas, iš kurio generuojamas Go ir TypeScript kodas |
| `web` | React ir TypeScript sąsaja |
| `docs` | [Eigos](docs/flows.md), [testavimo planas](docs/testing-plan.md), [pristatymo planas](docs/presentation-plan.md), [sprendimai](docs/decisions.md) |

API kodo generavimas po `openapi/openapi.yaml` pakeitimo: `go generate ./internal/api` ir `npm run generate` aplanke `web`.
