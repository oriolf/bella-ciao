const en = {
    global: {
        id_formats: {
            dni: "Spanish DNI",
            nie: "Spanish NIE",
            passport: "Spanish passport",
        }
    },
    pages: {
        faq: {
            title: "Frequently asked questions"
        },
        initialize: {
            error: "Already initialized"
        }
    },
    comp: {
        buttons: {
            delete_file: "Delete file",
            download_file: "Download file"
        },
        candidate_form: {
            form_name: "Add new candidate",
            general_error: "Could not add candidate",
            name: "Candidate's name",
            presentation: "Candidate's presentation",
            image: "Upload candidate's image",
            add_candidate: "Add candidate",
        },
        candidates_list: {
            delete: "Delete candidate",
            image: "Candidate {name} image",
            error: "Could not get candidates",
        },
        election_results: {
            name: "Name",
            points: "Points",
        },
        election_vote: {
            election: "Election",
            election_before: "The election will take place between {start} and {end}",
            election_started: 'The election has started and will be open until {end}',
            form_name: "Check vote",
            token: "Enter the identifier that was given when you voted",
            vote_confirmation: "You vote has been recorded, you can check for which candidates did you vote by providing the following identifier: {hash}",
            already_voted: "You have already voted",
            check_vote: "You can check your vote by providing the identifier that was shown to you when you voted:",
            candidates_list: "You voted for the following candidates:",
            candidates_range: "You must select at least {min} candidates and up to {max}.",
            candidates: "Candidates",
            selected_candidates: "Selected candidates",
            vote: "Vote",
            election_results: "The election has already ended, you can check the results here",
        },
        form: {
            field_required: "This field is required"
        },
        loading: {
            loading: "Loading...",
        },
        login_form: {
            login: "Log in",
            login_err: "The user or the password are invalid",
            unique_id_hint: "Enter user (identification number: DNI, NIE, etc.)",
            unique_id_err: "This field is required and must match a valid ID number format",
            password_hint: "Password",
            password_err: "This field is required and must be eight or more characters long",
            register: "Register",
            register_err: "The user is not a valid identification number (DNI, NIE...), the name or email are empty, or the passwords don't match",
            name_hint: "Enter your name",
            email_hint: "Enter your email",
            repeat_password_hint: "Repeat password",
            register_password_err: "This field is required and must match the other password field",
        },
        nav: {
            home: 'Home',
            candidates: 'Candidates',
            faq: 'FAQ',
            language: 'Language',
            logout: 'Log out',
        },
        role_admin: {
            pending_err: "Could not get users pending validation",
            validated_err: "Could not get validated users",
        },
        role_none: {
            upload_file: "Upload file",
            upload_file_err: "Could not upload file",
            description_hint: "Description of the file contents",
            none_notice: "You still have not been validated. Please review and solve the validators' messages and upload the required files",
            messages_title: "Messages from validators",
            messages_explanation: "Please review these messages, perform the requested action, and mark as solved when done.",
            messages_header: "Message",
            messages_err: "Could not obtain the validation messages",
            solved: "(already solved)",
            solve: "Mark as solved",
            files_title: "Uploaded files",
            files_explanation: "Upload the required files and remove those that are no longer necessary.",
        },
        role_validated: {
            validated: "You have been validated",
            uploaded_files: "Uploaded files",
        },
        uninitialized: {
            initialize: "Initialize",
            initialize_err: "Could not initialize site",
            unique_id_title: "Admin's unique ID",
            unique_id_hint: "Enter admin unique ID",
            unique_id_err: "This field is required and must match a valid ID number format",
            name_title: "Admin name",
            name_hint: "Enter the admin's name",
            email_title: "Admin email",
            email_hint: "Enter the admin's email",
            password_title: "Admin password",
            password_hint: "Password",
            password_err: "This field is required and must match the other password field",
            rpassword_title: "Repeat password",
            rpassword_hint: "Repeat password",
            election_name_title: "Election's name",
            election_name_hint: "Enter the election's name",
            start_title: "Election start",
            start_hint: "Date and time of election start",
            end_title: "Election end",
            end_hint: "Date and time of election end",
            count_method_title: "Vote's count method",
            count_method_hint: "Vote count method to use",
            min_candidates_title: "Mimimum candidates to vote for",
            min_candidates_hint: "The minimum number of candidates to select",
            min_candidates_err: "This field is required, must be greater than zero and less or equal to the maximum candidates",
            max_candidates_title: "Maximum candidates to vote for",
            max_candidates_hint: "The maximum number of candidates to select",
            max_candidates_err: "This field is required, must be greater than zero and greater or equal to the minimum candidates",
            id_formats_title: "Allowed identification formats",
            id_formats_hint: "The allowed identification formats for registration",
            id_formats_err: "This field is required, you should select at least one format",
        },
        user_files: {
            description: "Description",
            files_err: "Could not obtain the uploaded files",
        },
        users_pagination: {
            filter_user_id: "Filter by user identifier",
            add_message: "Add message",
            files: "Files",
            messages: "Messages",
            solved: "Already solved",
            not_solved: "Not solved yet",
            close: "Close",
        },
    },
}

export default en