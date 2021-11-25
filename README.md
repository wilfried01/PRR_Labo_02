# PRR 2021 / Labo 2

## Fonctionnalités de l'applicaiton
Dans cette version de l'applicaiton, nous avons mis en place l'utilisation de plusieurs sites pour réaliser les mêmes
fonctionnalités que celles du laboratoire 1. Afin de gérer la concurrence entre eux, les sites utilisent l'algorithme 
de Lamport.

## Lancement de l'application
Pour utiliser l'application il convient de procéder comme suite:

- Éditer le fichier configuration.json afin de spécifier 
  - Le nombre de serveurs
  - les hôtes et port pour les serveurs. `Autant que le nombre de serveurs annoncé`
  - le nombre de chambres
  - le nombre de jours


> **Attention. Si le fichier de configuration n'est pas bien éditer, l'application ne 
> fonctionnera pas. Un fichier est mis à disposition afin de fournir un exemple de 
> configuration**

- Se déplacer dans le dossier **_main_** et lancer le fichier main.go. Si cette opération réussie, vous devriez voir 
dans la console les messages indiquant que les serveurs ont été lancé avec succès.
- Se déplacer dans le dosser **_client_** et lancer le fichier main.go.
