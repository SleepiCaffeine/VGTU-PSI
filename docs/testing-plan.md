# Testavimo planas

Laukiami rezultatai paimti iš [I užduoties scenarijų](../I-uzduotis.md), o ne iš kodo.

## Dokumento scenarijai

Testai: `internal/service/engine/engine_test.go`.

| Scenarijus | Testas | Tikrinama |
| --- | --- | --- |
| Įprastas atvejis | `TestDesignScenariosBets` | `ACCEPTED`, be patvirtinimo, balansas 480 |
| Ribinis atvejis | `TestDesignScenariosBets` | `ACCEPTED_CAPPED`, statymas 10, `STOP_LOSS_CAP`, balansas 300, `STOP_LOSS_REACHED` |
| Konfliktas: vijimasis | `TestDesignScenariosBets` | `REJECTED`, `CHASING_LIMIT`, balansas nepakitęs, sesija baigta, įspėjimas |
| Trintis | `TestConfirmationVerification` | `CONFIRMATION_FAILED`, statymas neatliktas |
| Plano mažinimas | `TestPlanReduction` | Skyrimas 420, pažanga 16 % |
| Klaida | `TestSessionStartDecisions` | `DENIED`, `NO_CREDITS`, kito skyrimo data |

## Testų lygiai

| Lygis | Vieta | Paskirtis |
| --- | --- | --- |
| Variklis | `internal/service/engine` | Taisyklių eilės tvarka, trinties lygiai, ribos, apvalinimas |
| Žaidimai | `internal/service/game` | Išmokos su fiksuotu rezultatu, tikslus namų pranašumas |
| Paslaugos | `internal/service` | Atminties saugykla: vienkartinis patvirtinimas, savaitės keitimas, poveikio išsaugojimas |
| API | `internal/api/router_test.go` | Būsenų kodai, kliento konteksto tikrinimas, klaidų neatskleidimas |
| Integraciniai | `internal/api/isolation_integration_test.go` | Tikras MySQL: klientų izoliavimas, statymo išsaugojimas, lygiagretūs statymai |

## Klientų izoliavimas

`TestTenantIsolation` veikia per visą sistemą su tikra duomenų baze. Klientas A bando skaityti ir keisti kliento B būseną, sesijas, statymus ir įspėjimus.

- Kai kelyje nurodytas kitas klientas, laukiamas `403`.
- Kai per savo kelią naudojamas kito kliento sesijos ID, laukiamas `404`.
- Kiekvienu atveju tikrinama, kad atsakyme nėra kliento B balanso, o jo duomenys ir eilučių skaičius nepasikeitė.

## Paleidimas

```
go test ./...
docker compose up -d --wait
go test -tags=integration -p 1 ./...
```

Tas pats vykdoma CI: `.github/workflows/ci.yml`, vienetų testai „Ubuntu“, „Windows“ ir „macOS“ aplinkose.
