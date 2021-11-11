# PRR 2021 / Labo 1
## Description

## Serveur
Le serveur permet à plusieurs clients de se connecter simultanément. Il est accessible sur le port 3333 avec netcat.

Le serveur gère les demandes simultanées de plusieurs clients

Le serveur se trouve dans /server

## Client
Nous avons décidé que le client serait un simple telnet / netcat (nous recommandons netcat)

## Fonctionalités implémentées
- Réservation
- Affichage de l'occupation des chambres
- Obtention d'un numéro de chambre libre
- Gestion de plusieurs clients
- Gestion de la déconnexion soudaine

## Utilisation 
Afin de lancer le serveur, il suffit d'exécuter :
``` bash
go run server.go
```
Le serveur sera ainsi lancé avec les paramètres par défaut qui sont modifiables dans server.go

On peut également lancer le serveur avec :
``` bash
go run server.go <numberOfDays> <numberOfRooms> <debugMode (true/false)>
```
Ce qui permet de spécifier le nombre de jours et le nombre de chambres, ainsi que l'utilisation du mode debug qui va artificiellement ralentir le traitement des commandes, et permettant ainsi d'obtenir des situations où il y a de la concurrence entre les différentes commandes

## Choix d'implémentation
Les jours et numéros de chambres vont de 1 à N

Si une chambre est réservée pour un jour au jour N, elle sera affichée occupée pour le jour N, mais pas pour le jour N + 1.

Ce choix a été motivé par le fait que l'on réserve une chambre pour une nuit -> nuit de N à N + 1 et que la chambre peut être réutilisée pour la nuit de N + 1 à N + 2

L'input utilisateur est vérifié côté serveur et on notifiera l'utilisateur si une commnande n'existe pas où que les paramètres sont invalides. 

Dû à la nécessité de vérifier côté serveur que le jour + durée de séjour ne dépasse pas le nombre de jours maximum, il nous a paru inutile de créer un client qui validerait les input utilisateur mais devrait quand même valider certains paramètres côté serveur.

## Commandes implémentées
HELP : Affiche les différentes commandes

RESERVE <day> <room> <duration> : Permet de réserver la chambre passée en paramètre, à partir du jour passé en paramètre et pour une durée passée en paramètre

DISPLAY <day> : Permet d'afficher l'occupation des chambres un jour donné

GETFREE <day> <duration> : Permet d'obtenir le numéro d'une chambre libre à partir du jour passé en paramètre pour un durée passée en paramètre

## Test
Il faut re-lancer le serveur avant chaque exécution des tests
``` bash
go run server.go
```
Pour tester il suffit d'utiliser la commande suivante sur le package server:

``` bash
go test -v
```

## Test accès Concurrent
Les tests qui ont été réalisés sont les suivants :
- Tester que le serveur puisse gérer 2 réservations de clients pour des chambres différentes
- Tester que lorsque 2 clients font 2 réservations pour une même chambre pour des jours qui se chevauchent, une seule est acceptée
- Tester que la commande display affiche les mêmes informations (chambre occupée ou libre) pour 2 utilisateurs

Les différents tests ont été réalisés en utilisant le mode debug afin d'avoir plusieurs commandes qui s'exécutent de manière concurrente

## Equipe
- Manuel Carvalho
- Noah Fusi
- Karel Ngueukam
