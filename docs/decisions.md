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
| Kaip skaičiuojami vijimosi signalai? | Per visą plano savaitę, kad naujos sesijos pradėjimas jų nenunulintų. Tai keičia I užduoties sesijos lygio skaičiavimą |
| Kas vyksta po trečio signalo? | Sesija baigiama ir iki kito skyrimo naujos sesijos nepradedamos (`CHASING_LOCKOUT`). I užduotis numato tik sesijos pabaigą, bet be užrakinimo taisyklę būtų galima apeiti |
| Ar pertrauką galima praleisti baigus sesiją? | Ne, kol pertrauka nesibaigė, nauja sesija nepradedama (`BREAK_ACTIVE`) |
| Ar užrakinimą gali panaikinti prižiūrėtojas? | Ne šiame etape, patvirtinimų eiga neįgyvendinta |
| Ar statymas priimamas po 1 ir 2 signalo? | Taip, po jo prasideda privaloma pertrauka |
| Kada informuojamas prižiūrėtojas? | Po kiekvieno vijimosi signalo (`CHASING_SIGNAL`), o trečias savaitės signalas dar ir baigia sesiją (`CHASING_LIMIT`). I užduotis numato pranešimą tik po trečio, bet prižiūrėtojas turi matyti ir ankstyvus ženklus |
| Kokia statymo reikšmė lyginama vijimuisi? | Prašomas statymas, iki sumažinimo |
| Kaip nustatomas trinties lygis? | Savaitės juosta + statymų juosta + šios savaitės signalai + greitų statymų žymos, daugiausia 4. Slenksčiai yra `RuleSettings` |
| Ar ribojami labai greiti statymai? | Taip, papildoma taisyklė, kurios I užduotyje nėra. Jei `RapidBetLimit` statymų iš eilės skiria mažiau nei `RapidBetSeconds`, statymas atmetamas (`RAPID_BETTING`), prasideda pertrauka, informuojamas prižiūrėtojas, trintis padidėja vienu lygiu |
| Ar patvirtinti galima iš karto? | Ne, serveris kiekvieną patvirtinimą priima ne anksčiau kaip po 1 s, o 3 lygyje po `FrictionWaitSeconds` |
| Kas yra „reguliariai“? | Kas `ConfirmEveryBets` statymų |
| Koks sakinys prašomas? | Su galutiniu, jau sumažintu statymu ir žaidimo namų pranašumu |
| Ar nesėkmingas patvirtinimas jį panaikina? | Ne, panaudotu pažymimas tik priimtas statymas |
| Ar praleista savaitė laikoma švaria? | Taip, joje nebuvo vijimosi |
| Laiko juosta | Plano savaitės skaičiuojamos Europe/Vilnius, duomenys saugomi UTC |
| Kaip atsiranda nauja savaitė? | Tingiai, per pirmą užklausą po pirmadienio 00:00 |

## Neįgyvendinta šiame etape

Pagal II užduoties apimtį neįgyvendinta: lošimo automatai, „Plinko“, „Blackjack“, pranešimai apie atkrytį, ribų keitimo eiga, poros susiejimas ir atsiejimas, prisijungimas, prižiūrėtojo veiksmų žurnalas. Taip pat 40 užklausų prieigos patikra ir modifikuojamumo bandymas numatyti III etapui.
