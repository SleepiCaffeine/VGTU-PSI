# Pristatymo planas

Pasiruošimas: paleista „Docker Desktop“, įvykdyta `go run ./cmd/dev -reset`, atidaryta sąsaja, šalia terminalas.

| Eilė | Trukmė | Ką rodyti | Ką pasakyti |
| --- | --- | --- | --- |
| 1 | 1 min | Problema, dvi veiksenos | Pagrindinis modulis yra taisyklių variklis, žaidimas tik jį naudoja |
| 2 | 2 min | Poros A statymas 20, `ACCEPTED` | Serveris nustato rezultatą, sprendimas turi priežastį |
| 3 | 1 min | Savaitės keitimas, 420 ir 16 % | Pažanga vienoda abiem naudotojams |
| 4 | 2 min | Poros B statymas 40, `CHASING_LIMIT`, įspėjimas | Konfliktas tarp vijimosi ir statymo ribų sprendžiamas eilės tvarka |
| 5 | 1 min | Neteisingas sakinys, `CONFIRMATION_FAILED` | Patvirtinimą tikrina serveris, jis vienkartinis |
| 6 | 1 min | `curl` su `403` ir `404` | Atskyrimas yra ne sąsajoje: tikrinimas tarpinėje programoje ir `inTenant` saugykloje |
| 7 | 1 min | `go test -tags=integration -p 1 ./...` | Scenarijai paimti iš I užduoties |
| 8 | 2 min | Kodo struktūra, `engine/registry.go` | Priklausomybės nukreiptos į vidų |

## Galimas pakeitimas gyvai

Taisyklė „ne daugiau kaip 3 sesijos per dieną“:

1. Naujas failas `internal/service/engine/rule_daily_sessions.go` su `Rule` realizacija.
2. Viena eilutė `internal/service/engine/registry.go` sesijos pradžios sąraše.
3. Naujas laukas `SessionsToday` `engine.Input` ir jo užpildymas `SessionService`, nes reikia naujų duomenų.

Esami taisyklių failai nekeičiami.

Naujas žaidimas: failas `internal/service/game/keno.go` su `Game` realizacija ir vienas argumentas `game.Default()`.

## Galimi klausimai

| Klausimas | Atsakymas |
| --- | --- |
| Kodėl 409, o ne 200, kai statymas atmestas? | Veiksmas neatliktas dėl būsenos konflikto, atsakyme tas pats sprendimo objektas |
| Kaip išvengiama dvigubo išleidimo? | Plano eilutė užrakinama transakcijoje (`SELECT ... FOR UPDATE`) |
| Ar `X-Tenant-ID` saugus? | Ne, tai laikina tapatybės vieta. Prisijungimas numatytas III etapui |
| Kodėl 404, o ne 403, kai sesija kito kliento? | Saugykla jos nemato, todėl neatskleidžiama, kad toks įrašas egzistuoja |
