- Fer la pàgina inicial per als tres tipus d'usuari (HomePageContentNone, Validated i Admin).
- Els usuaris no validats han de poder tindre "comentaris" de l'admin que els ha rebutjat, i els han de poder marcar com a resolts.
- A l'inicialitzar i al registrar-se correctament hauria de redirigir a alguna altra pàgina.

Idees a considerar
------------------


- Dockeritzada, que es puga fer un docker run test i un docker run i que tot funcione bé.
- Per tant no necessita ni tindre un apache, ni un mysql, ni res, tot està dins del docker.
- Registre super fàcil, pots pujar els documents que vulgues i entrar només registrar-te.
- Quan entres, pàgina simple que et diu el teu status (si estàs aprovat o no). Si no t'aproven, et mostra el perquè, i pots subsanar la documentació. Tens una llista de les accions dutes a terme.
- Sense circumscripcions ni històries rares, només dos tipus d'usuaris, els que creen votacions, aproven usuaris, etc. I els que només voten.
- Només una votació per instal·lació. Com és tan fàcil instal·lar, no cal més.
- Configurable el tipus d'identificació que es permet. Dins de la llista d'opcions (DNI, NIE...) es marquen les que vols, que son les que permet fer servir.
- Configurable el tipus de recompte, el mínim i el màxim de vots, etc.
- Configurables els candidats, poder-los posar nom, descripció i una foto, i tindre una pàgina guai de candidats i tal.
- Fer una bona pàgina de FAQs.
- Fer una bona landing page, amb enllaços (o parts de) els candidats, les FAQ, etc.
- Permetre registrar una persona directament per a votació presencial.
- Una pàgina amb les estadistiques de participació.
- To inform users that they have been validated without using email: https://medium.com/finizens-engineering/usando-push-notifications-en-tu-web-7e6711e9070e
  https://justmarkup.com/articles/2017-02-07-implementing-push-notifications/
  https://github.com/web-push-libs
