# PRR 2021 / Labo 2

## Fonctionnalités de l'applicaiton
Dans cette version de l'applicaiton, nous avons mis en place l'utilisation de plusieurs sites pour réaliser les mêmes
fonctionnalités que celles du laboratoire 1. Afin de gérer la concurrence entre eux, les sites utilisent l'algorithme 
de Lamport.

## Architecture
Le dossier **_main_** contient un code permettant de lancer les serveurs suivant le fichier de configuration.

Le dossier **_server_** contient le code serveur qui s'occupe de l'exclusion mutuelle avec Lamport

Le dossier **_server/hotel_** contient le code du service de réservation (labo 1). Ce dernier a été légèrement modifié afin de l'intégrer dans le serveur

Le dossier **_client_** contient le code du client, il nous a paru nécessaire de réaliser un client afin de faciliter les tests

## Lancement de l'application
Pour utiliser l'application il convient de procéder comme suite:

- Éditer le fichier configuration.json afin de spécifier 
  - Le nombre de serveurs
  - les hôtes et port pour les serveurs. `Autant que le nombre de serveurs annoncé`
  - le nombre de chambres
  - le nombre de jours


> **Attention. Si le fichier de configuration n'est pas bien édité, l'application ne 
> fonctionnera pas. Un fichier est mis à disposition afin de fournir un exemple de 
> configuration**

- Se déplacer dans le dossier **_main_** et lancer le fichier main.go. Si cette opération réussie, vous devriez voir 
dans la console les messages indiquant que les serveurs ont été lancé avec succès.
- Se déplacer dans le dosser **_client_** et lancer le fichier main.go.

## Test

### Tests Unitaires

Il faut re-lancer les serveurs avant chaque exécution des tests dans le package main:
``` bash
go run main.go
```
Pour tester il suffit d'utiliser la commande suivante sur le package server:

``` bash
go test -v
```
### Scénarios de flux d'Exécution

Il faut a nouveau re-lancer dans le package main(conseiller d'utiliser le mode debug dans les scénarios):

Normal:
``` bash
go run main.go
```
Debug:
``` bash
go run main.go DEBUG
```
il est aussi conseillé d'ouvrir deux client-consoles pour testes les divers scénarios :

``` bash
go run main.go  (dans le package client)
```

De plus, les logs on le formât suivant :

(server)(Type De Message)(Horloge de Lamport)(RECEIVED FROM|SENT TO)(server)


#### Scénario 1 - Une seule réservation d'un client

-Objectif: voir le fonctionnement général de l'algorithme de Lamport

-Résultats attendus:

``` bash

RECEIVED RESERVE 1 3 2 MAXIME
SENDING REQUESTS FROM SERVER : 0
1: REQ 1 RECEIVED FROM 0
1: ACK 2 SENT TO 0
3: REQ 1 RECEIVED FROM 0
2: REQ 1 RECEIVED FROM 0
2: ACK 2 SENT TO 0
3: ACK 2 SENT TO 0
4: REQ 1 RECEIVED FROM 0
4: ACK 2 SENT TO 0
0: ACK 2 RECEIVED FROM 1
0: ACK 2 RECEIVED FROM 3
0: ACK 2 RECEIVED FROM 2
0: ACK 2 RECEIVED FROM 4
Server 0 entering SC
Sending LOCALREL
Server 0 leaving SC
1: REL 7 RECEIVED FROM 0
2: REL 7 RECEIVED FROM 0
3: REL 7 RECEIVED FROM 0
4: REL 7 RECEIVED FROM 0

```

Dans les résultats attendus on peut bien voir le bon fonctionnement de l'algorithme de Lamport, le server lance un message a toutes les autres servers du type request, 
et il rentre seulement dans la section critique quand il reçoit tout les ACK avec des timestamps supérieurs à la REQ lancer par le server 0 (2>1).
À la fin de l'utilisation de la section critique le serveur lance un message du type release aux autres servers.

#### Scénario 2 - Deux réservations concurrentes de la SC

-Objectif: voir le fonctionnement général de l'algorithme de Lamport dans une situation de concurrence de la SC

-Résultats attendus:

``` bash

RECEIVED RESERVE 1 3 2 MAXIME
SENDING REQUESTS FROM SERVER : 0
RECEIVED RESERVE 1 6 2 ANDRE
SENDING REQUESTS FROM SERVER : 1
2: REQ 1 RECEIVED FROM 0
2: ACK 2 SENT TO 0
3: REQ 1 RECEIVED FROM 0
3: ACK 2 SENT TO 0
4: REQ 1 RECEIVED FROM 0
4: ACK 2 SENT TO 0
3: REQ 1 RECEIVED FROM 1
3: ACK 3 SENT TO 1
0: ACK 2 RECEIVED FROM 2
4: REQ 1 RECEIVED FROM 1
4: ACK 3 SENT TO 1
2: REQ 1 RECEIVED FROM 1
2: ACK 3 SENT TO 1
1: REQ 1 RECEIVED FROM 0
1: ACK 2 SENT TO 0
0: ACK 2 RECEIVED FROM 3
1: ACK 3 RECEIVED FROM 3
0: ACK 2 RECEIVED FROM 4
1: ACK 3 RECEIVED FROM 4
Server 0 entering SC
0: REQ 1 RECEIVED FROM 1
0: ACK 6 SENT TO 1
1: ACK 3 RECEIVED FROM 2
0: ACK 2 RECEIVED FROM 1
1: ACK 6 RECEIVED FROM 0
Sending LOCALREL
Server 0 leaving SC
Server 1 entering SC
1: REL 8 RECEIVED FROM 0
2: REL 8 RECEIVED FROM 0
3: REL 8 RECEIVED FROM 0
4: REL 8 RECEIVED FROM 0
Sending LOCALREL
Server 1 leaving SC
0: REL 10 RECEIVED FROM 1
2: REL 10 RECEIVED FROM 1
3: REL 10 RECEIVED FROM 1
4: REL 10 RECEIVED FROM 1

```

Dans les résultats obtenus on a bien le server 1 qui rentre dans la SC après la fin du server 0. 
On a bien une exclusion mutuelle pour accéder a la section critique.

#### Scénario 3- Deux réservations concurrentes de la SC et de cohérence de donnés

-Objectif: voir le fonctionnement général de l'algorithme de Lamport dans une situation de concurrence de la SC et de cohérence de donnés

-Résultats attendus:

``` bash

RECEIVED RESERVE 1 3 2 MAXIME
SENDING REQUESTS FROM SERVER : 0
RECEIVED RESERVE 1 3 2 ANDRE
SENDING REQUESTS FROM SERVER : 1
2: REQ 1 RECEIVED FROM 0
2: ACK 2 SENT TO 0
3: REQ 1 RECEIVED FROM 0
3: ACK 2 SENT TO 0
4: REQ 1 RECEIVED FROM 0
4: ACK 2 SENT TO 0
3: REQ 1 RECEIVED FROM 1
3: ACK 3 SENT TO 1
0: ACK 2 RECEIVED FROM 2
4: REQ 1 RECEIVED FROM 1
4: ACK 3 SENT TO 1
2: REQ 1 RECEIVED FROM 1
2: ACK 3 SENT TO 1
1: REQ 1 RECEIVED FROM 0
1: ACK 2 SENT TO 0
0: ACK 2 RECEIVED FROM 3
1: ACK 3 RECEIVED FROM 3
0: ACK 2 RECEIVED FROM 4
1: ACK 3 RECEIVED FROM 4
Server 0 entering SC
0: REQ 1 RECEIVED FROM 1
0: ACK 6 SENT TO 1
1: ACK 3 RECEIVED FROM 2
0: ACK 2 RECEIVED FROM 1
1: ACK 6 RECEIVED FROM 0
Sending LOCALREL
Server 0 leaving SC
Server 1 entering SC
1: REL 8 RECEIVED FROM 0
2: REL 8 RECEIVED FROM 0
3: REL 8 RECEIVED FROM 0
4: REL 8 RECEIVED FROM 0
Sending LOCALREL
Server 1 leaving SC
0: REL 10 RECEIVED FROM 1
2: REL 10 RECEIVED FROM 1
3: REL 10 RECEIVED FROM 1
4: REL 10 RECEIVED FROM 1

```
Les mêmes logs sont obtenus, néanmoins, la réservation de MAXIME va être faite lorsque celle
de ANDRE ne vas pas être possible. Ceci prouve la cohérence des et la mise-a-jour des données dans cette architecture pluri-serveurs.


