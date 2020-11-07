const ca = {
    global: {
        id_formats: {
            dni: "DNI espanyol",
            nie: "NIE espanyol",
            passport: "Passaport espanyol",
        }
    },
    pages: {
        faq: {
            title: "Preguntes freqüents"
        },
        initialize: {
            error: "Ja s'ha inicialitzat"
        }
    },
    comp: {
        buttons: {
            delete_file: "Esborra arxiu",
            download_file: "Descarrega arxiu"
        },
        candidate_form: {
            form_name: "Afegeix candidat nou",
            general_error: "No s'ha pogut afegir el candidat",
            name: "Nom del candidat",
            presentation: "Presentació del candidat",
            image: "Puja la imatge del candidat",
            add_candidate: "Afegeix candidat",
        },
        candidates_list: {
            delete: "Esborra candidat",
            image: "Imatge del candidat {name}",
            error: "No s'han pogut obtindre els candidats",
        },
        election_results: {
            name: "Nom",
            points: "Punts",
        },
        election_vote: {
            election: "Votació",
            election_before: "La votació començarà el {start} i durarà fins {end}",
            election_started: 'La votació ha començat i durarà fins el {end}',
            form_name: "Comprova el vot",
            token: "Introdueix l'identificador que et van mostrar quan quan vas votar",
            vote_confirmation: "El vot s'ha registrat, pots comprovar per quins candidats vas votar proporcionant el següent identificador: {hash}",
            already_voted: "Ja has votat",
            check_vote: "Pots comprovar el teu vot proporcionant l'identificador que se't va mostrar quan vas votar:",
            candidates_list: "Vas votar pels candidats següents:",
            candidates_range: "Has de seleccionar entre {min} i {max} candidats.",
            candidates: "Candidats",
            selected_candidates: "Candidats seleccionats",
            vote: "Vota",
            election_results: "La votació ja ha acabat, pots comprovar els resultats ací",
        },
        form: {
            field_required: "Aquest camp és requerit"
        },
        loading: {
            loading: "Carregant...",
        },
        login_form: {
            login: "Entra",
            login_err: "L'usuari o la contrasenya no son vàlids",
            unique_id_hint: "Introdueix l'usuari (número identificatiu: DNI, NIE...)",
            unique_id_err: "Aquest camp és requerit i ha de tindre un format de número d'identificació vàlid",
            password_hint: "Contrasenya",
            password_err: "Aquest camp és requerit i ha de tindre vuit caràcters o més",
            register: "Registrar-se",
            register_err: "L'usuari no és un número d'identificació vàlid (DNI, NIE...), el nom o el correu estan buits, o les contrasenyes no coincideixen",
            name_hint: "Introdueix el teu nom",
            email_hint: "Introdueix el teu correu electrònic",
            repeat_password_hint: "Repeteix la contrasenya",
            register_password_err: "Aquest camp és requerit i ha de coincidir amb l'altre camp de contrasenya",
        },
        nav: {
            home: 'Inici',
            candidates: 'Candidats',
            faq: 'FAQ',
            language: 'Idioma',
            logout: 'Surt',
        },
        role_admin: {
            pending_err: "No s'han pogut obtenir els usuaris pendents de validació",
            validated_err: "No s'han pogut obtenir els usuaris validats",
        },
        role_none: {
            upload_file: "Puja l'arxiu",
            upload_file_err: "No s'ha pogut pujar l'arxiu",
            description_hint: "Descripció del contingut de l'arxiu",
            none_notice: "Encara no t'han validat. Per favor, revisa i resol els missatges dels validadors i puja els arxius requerits",
            messages_title: "Missatges dels validadors",
            messages_explanation: "Per favor, revisa aquests missatges, fes l'acció que es demana, i marca com a resolt quan ho hages fet.",
            messages_header: "Missatge",
            messages_err: "No s'han pogut obtindre els missatges de validació",
            solved: "(ja resolt)",
            solve: "Marca com a resolt",
            files_title: "Fitxers pujats",
            files_explanation: "Puja els arxius requerits i elimina aquells que ja no son necessaris.",
        },
        role_validated: {
            validated: "Ja has sigut validat",
            uploaded_files: "Fitxers pujats",
        },
        uninitialized: {
            initialize: "Inicialitza",
            initialize_err: "No s'ha pogut initialitzar el lloc",
            unique_id_title: "Identificador únic de l'administrador",
            unique_id_hint: "Introdueix l'identificador únic de l'administrador",
            unique_id_err: "Aquest camp és requerit i ha de tindre un format de número d'identificació vàlid",
            name_title: "Nom de l'administrador",
            name_hint: "Introdueix el nom de l'administrador",
            email_title: "Correu electrònic de l'administrador",
            email_hint: "Introdueix el correu electrònic de l'administrador",
            password_title: "Contrasenya de l'administrador",
            password_hint: "Contrasenya",
            password_err: "Aquest camp és requerit i ha de coincidir amb l'altre camp de contrasenya",
            rpassword_title: "Repeteix la contrasenya",
            rpassword_hint: "Repeteix la contrasenya",
            election_name_title: "Nom de la votació",
            election_name_hint: "Introdueix el nom de la votació",
            start_title: "Inici de la votació",
            start_hint: "Hora i dia quan comença la votació",
            end_title: "Fi de la votació",
            end_hint: "Hora i dia quan acaba la votació",
            count_method_title: "Mètode de comptatge de vots",
            count_method_hint: "Mètode de comptatge de vots a fer servir",
            min_candidates_title: "Mínim de candidats pels quals cal votar",
            min_candidates_hint: "El mínim de candidats que cal seleccionar",
            min_candidates_err: "Aquest camp és requerit, ha de ser major de zero i menor o igual que el màxim de candidats",
            max_candidates_title: "Màxim de candidats pels quals es pot votar",
            max_candidates_hint: "El màxim de candidats que es pot seleccionar",
            max_candidates_err: "Aquest camp és requerit, ha de ser major de zero i major o igual que el mínim de candidats",
            id_formats_title: "Formats d'identificació permesos",
            id_formats_hint: "Els formats d'identificació que es permeten per registrar-se",
            id_formats_err: "Aquest camp és requerit, has de seleccionar com a mínim un format",
        },
        user_files: {
            description: "Descripció",
            files_err: "No s'han pogut obtindre els arxius",
        },
        users_pagination: {
            filter_user_id: "Filtra per identificador d'usuari",
            add_message: "Afegeix missatge",
            files: "Arxius",
            messages: "Missatges",
            solved: "Resolt",
            not_solved: "No resolt",
            close: "Tanca",
        },
    },
}

export default ca