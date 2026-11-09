# Sprendimai

## Technologijos

| Sprendimas | Priežastis |
| --- | --- |
| Klientas (tenant) yra lošiančio asmens ir prižiūrėtojo pora | Atitinka I užduoties poros narystės tikrinimą |
| MySQL per `docker compose` | Atitinka projektavimo dokumentą |
| Sąsaja įterpiama į Go programą | Vienas adresas ir paprastas paleidimas |
| Paleidimas per `go run ./cmd/dev` | „Windows“ neturi `make`, o Go jau būtinas |
| Lentelės kuriamos `AutoMigrate` | Pakanka prototipui |

## Taisyklių aiškinimas

Vietos, kuriose I užduotis nenurodo tikslaus elgesio.

| Klausimas | Sprendimas |
| --- | --- |
| Kaip skaičiuojami vijimosi signalai? | Atskirai kiekvienai sesijai |
| Ar statymas priimamas po 1 ir 2 signalo? | Taip, po jo prasideda privaloma pertrauka |
| Kokia statymo reikšmė lyginama vijimuisi? | Prašomas statymas, iki sumažinimo |
| Kaip nustatomas trinties lygis? | Savaitės juosta + statymų juosta + signalai, daugiausia 4. Slenksčiai yra `RuleSettings` |
| Kas yra „reguliariai“? | Kas `ConfirmEveryBets` statymų |
| Koks sakinys prašomas? | Su galutiniu, jau sumažintu statymu ir žaidimo namų pranašumu |
| Ar nesėkmingas patvirtinimas jį panaikina? | Ne, panaudotu pažymimas tik priimtas statymas |
| Ar praleista savaitė laikoma švaria? | Taip, joje nebuvo vijimosi |
| Laiko juosta | Plano savaitės skaičiuojamos Europe/Vilnius, duomenys saugomi UTC |
| Kaip atsiranda nauja savaitė? | Tingiai, per pirmą užklausą po pirmadienio 00:00 |

## Neįgyvendinta šiame etape

Pagal II užduoties apimtį neįgyvendinta: lošimo automatai, „Plinko“, „Blackjack“, pranešimai apie atkrytį, ribų keitimo eiga, poros susiejimas ir atsiejimas, prisijungimas, prižiūrėtojo veiksmų žurnalas. Taip pat 40 užklausų prieigos patikra ir modifikuojamumo bandymas numatyti III etapui.
