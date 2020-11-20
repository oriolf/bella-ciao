const es = {
    global: {
        id_formats: {
            dni: "DNI español",
            nie: "NIE español",
            passport: "Pasaporte español",
        }
    },
    pages: {
        faq: {
            title: "Preguntas frecuentes"
        },
        initialize: {
            error: "Ya se ha inicializado"
        }
    },
    comp: {
        buttons: {
            delete_file: "Borra el archivo",
            download_file: "Descarga el archivo"
        },
        candidate_form: {
            form_name: "Añade nuevo candidato",
            general_error: "No se ha podido añadir al candidato",
            name: "Nombre del candidato",
            presentation: "Presentación del candidato",
            image: "Sube la imagen del candidato",
            add_candidate: "Añade candidato",
        },
        candidates_list: {
            delete: "Borra candidato",
            image: "Imagen del candidato {name}",
            error: "No se han podido obtener los candidatos",
        },
        election_results: {
            name: "Nombre",
            points: "Puntos",
        },
        election_vote: {
            election: "Votación",
            election_before: "La votación empezará el {start} i durará hasta el {end}",
            election_started: 'La votación ha empezado y durará hasta el {end}',
            form_name: "Comprueba el voto",
            token: "Introduce el identificador que se te mostró cuando votaste",
            vote_confirmation: "El voto se ha registrado, puedes comprobar por qué candidatos votaste proporcionando el siguiente identificador: {hash}",
            already_voted: "Ya has votado",
            check_vote: "Puedes comprobar tu voto proporcionando el identificador que se te mostró cuando votaste:",
            candidates_list: "Votaste por los siguientes candidatos:",
            candidates_range: "Tienes que seleccionar entre {min} y {max} candidatos.",
            candidates: "Candidatos",
            selected_candidates: "Candidatos seleccionados",
            vote: "Vota",
            election_results: "La votación ha terminado, puedes comprobar los resultados aquí",
        },
        form: {
            field_required: "Este campo es requerido"
        },
        loading: {
            loading: "Cargando...",
        },
        login_form: {
            login: "Entra",
            login_err: "El usuario o la contraseña no son válidos",
            unique_id_hint: "Introduce el usuario (número identificativo: DNI, NIE...)",
            unique_id_err: "Este campo es requerido y tiene que tener un formato de número de identificación válido",
            password_hint: "Contraseña",
            password_err: "Este campo es requerido y tiene que tener ocho caracteres o más",
            register: "Registrarse",
            register_err: "El usuario no es un numero de identificación válido (DNI, NIE...), el nombre o el correo estan vacíos, o las contraseñas no coinciden",
            name_hint: "Introduce tu nombre",
            email_hint: "Introduce tu correo electrónico",
            repeat_password_hint: "Repite la contraseña",
            register_password_err: "Este campo es requerido y tiene que coincidir con el otro campo de contraseña",
        },
        nav: {
            home: 'Inicio',
            candidates: 'Candidatos',
            faq: 'FAQ',
            language: 'Idioma',
            logout: 'Salir',
        },
        role_admin: {
            pending_err: "No se han podido obtener los usuarios pendientes de validación",
            validated_err: "No se han podido obtener los usuarios validados",
        },
        role_none: {
            upload_file: "Sube el fichero",
            upload_file_err: "No se ha podido subir el fichero",
            description_hint: "Descripción del contenido del fichero",
            none_notice: "Aún no has sido validado. Por favor, revisa y resuelve los mensajes de los validadores y sube los ficheros requeridos",
            messages_title: "Mensajes de los validadores",
            messages_explanation: "Por favor, revisa estos mensajes, realiza la acción requerida, y marca como resuelto cuando lo hayas hecho",
            messages_header: "Mensaje",
            messages_err: "No se han podido obtener los mensajes de validación",
            solved: "(ya resuelto)",
            solve: "Marca como resuelto",
            files_title: "Ficheros subidos",
            files_explanation: "Sube los ficheros requeridos y elimina los que ya no sean necesarios.",
        },
        role_validated: {
            validated: "Ya te han validado",
            uploaded_files: "Ficheros subidos",
        },
        uninitialized: {
            initialize: "Inicializa",
            initialize_err: "No se ha podido inicializar el sitio",
            unique_id_title: "Identificador único del administrador",
            unique_id_hint: "Introduce el identificador único del administrador",
            unique_id_err: "Este campo es requerido y tiene que tener un formato de numero de identificación válido",
            name_title: "Nombre del administrador",
            name_hint: "Introduce el nombre del administrador",
            email_title: "Correo electrónico del administrador",
            email_hint: "Introduce el correo electrónico del administrador",
            password_title: "Contraseña del administrador",
            password_hint: "Contraseña",
            password_err: "Este campo es requerido y tiene que coincidir con el otro campo de contraseña",
            rpassword_title: "Repite la contraseña",
            rpassword_hint: "Repite la contraseña",
            election_name_title: "Nombre de la votación",
            election_name_hint: "Introduce el nombre de la votación",
            start_title: "Inicio de la votación",
            start_hint: "Hora y día en el que empieza la votación",
            end_title: "Fin de la votación",
            end_hint: "Hora y día en el que acaba la votación",
            count_method_title: "Método de recuento de votos",
            count_method_hint: "Método de recuento de votos a utilizar",
            min_candidates_title: "Mínimo de candidatos por los que hay que votar",
            min_candidates_hint: "Mínimo de candidatos que hay que seleccionar",
            min_candidates_err: "Este campo es requerido, tiene que ser mayor que cero y menor o igual al máximo de candidatos",
            max_candidates_title: "Máximo de candidatos por los que se puede votar",
            max_candidates_hint: "Máximo de candidatos que se pueden seleccionarEl màxim de candidats que es pot seleccionar",
            max_candidates_err: "Este campo es requerido, tiene que ser mayor que cero y mayor o igual que el mínimo de candidatos",
            id_formats_title: "Formatos de identificación permitidos",
            id_formats_hint: "Los formatos de identificación que permiten registrarse",
            id_formats_err: "Este campo es requerido, tienes que seleccionar al menos un formato",
        },
        user_files: {
            description: "Descripción",
            files_err: "No se han podido obtener los ficheros",
        },
        users_pagination: {
            filter_user_id: "Filtra por identificador de usuario",
            add_message: "Añade mensaje",
            files: "Ficheros",
            messages: "Mensajes",
            solved: "Resuelto",
            not_solved: "No resuelto",
            close: "Cierra",
            validate: "Valida",
        },
    },
}

export default es