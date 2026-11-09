# Veikimo eigos

## Statymo sprendimas

```mermaid
flowchart TD
    A["POST .../bets"] --> B["Kliento konteksto tikrinimas"]
    B -->|nesutampa| X403["403"]
    B --> C["Žaidimas ir pasirinkimas galiojantys?"]
    C -->|ne| X400["400"]
    C --> D["Transakcija: užrakinamas planas, pritaikomas savaitės keitimas"]
    D --> E["Sesija ieškoma tik kliento duomenyse"]
    E -->|nerasta| X404["404"]
    E --> R1["Sesija aktyvi, laiko limitas"]
    R1 -->|atmesta| STOP["REJECTED + priežastis"]
    R1 --> R2["Privaloma pertrauka"]
    R2 -->|atmesta| STOP
    R2 --> R3["Vijimasis"]
    R3 -->|trečias signalas| STOP
    R3 --> R4["Statymo ribos"]
    R4 -->|nieko neliko| STOP
    R4 --> R5["Trintis ir patvirtinimas"]
    R5 -->|reikia| CONF["CONFIRMATION_REQUIRED arba CONFIRMATION_FAILED"]
    R5 --> PLAY["Žaidimas nustato rezultatą"]
    PLAY --> SETTLE["Balansas, sesija, statymo įrašas"]
    SETTLE --> OK["200 ACCEPTED arba ACCEPTED_CAPPED"]
    STOP --> X409["409 su sprendimu"]
    CONF --> X409
```

## Patvirtinimas

```mermaid
sequenceDiagram
    participant N as Naršyklė
    participant S as Serveris
    N->>S: Statymas be patvirtinimo
    S->>S: Sukuriamas vienkartinis įrašas su sakiniu ir statymo dydžiu
    S-->>N: 409 CONFIRMATION_REQUIRED, ID ir sakinys
    N->>S: Statymas su ID ir įvestu sakiniu
    alt Neteisingas sakinys, per anksti arba kitas statymas
        S-->>N: 409 CONFIRMATION_FAILED, balansas nepakitęs
    else Teisingas
        S->>S: Įrašas pažymimas panaudotu, raundas sužaidžiamas
        S-->>N: 200 ACCEPTED
    end
```

Sprendimo poveikis (vijimosi signalas, pertrauka) atmetamas, kai statymas sustoja ties patvirtinimu, todėl pakartotas statymas signalo neskaičiuoja dukart.

## Savaitės keitimas

```mermaid
flowchart TD
    A["Bet kuri būsenos, sesijos ar statymo užklausa"] --> B["Užrakinamas planas"]
    B --> C{"Dabar >= kito pirmadienio 00:00?"}
    C -->|ne| D["Tęsiama"]
    C -->|taip| E{"Tą savaitę buvo sesija, baigta dėl vijimosi?"}
    E -->|ne, švari| F["Skyrimas × 0,85, žemyn iki dešimčių"]
    E -->|taip| G["Skyrimas nesikeičia"]
    F --> H["Balansas = skyrimas, savaitė + 1"]
    G --> H
    H --> C
```

## Kliento duomenų atskyrimas

```mermaid
flowchart LR
    A["Užklausa /tenants/{id}/..."] --> B{"id = X-Tenant-ID?"}
    B -->|ne| C["403"]
    B -->|taip| D["Paslauga"]
    D --> E["Saugykla: kiekviena užklausa su tenant_id"]
    E -->|kito kliento įrašas| F["404, nieko nerašoma"]
    E --> G["Duomenys"]
```
