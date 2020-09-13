<script>
  import Form from "./Form.svelte";
  import { COUNT_METHODS, ID_FORMATS } from "../constants.js";
  import { validateArrayLengthPositive } from "../util.ts";

  function siteInitialized() {
    window.location = "/";
  }

  function composeTime(v, field) {
    return new Date(
      v[field + "_date"] + " " + v[field + "_time"]
    ).toISOString();
  }

  function groupParameters(v) {
    return {
      admin: {
        name: v.name,
        unique_id: v.unique_id,
        email: v.email,
        password: v.password,
      },
      election: {
        name: v.election_name,
        start: composeTime(v, "start"),
        end: composeTime(v, "end"),
        count_method: v.count_method,
        min_candidates: +v.min_candidates,
        max_candidates: +v.max_candidates,
      },
      config: {
        id_formats: v.id_formats,
      },
    };
  }

  let candidateForm = {
    name: "Initialize",
    url: "/api/initialize",
    generalError: "Could not initialize site",
    jsonFunc: groupParameters,
    fields: [
      // Admin fields
      {
        name: "unique_id",
        title: "Admin's unique ID",
        hint: "Enter admin unique ID",
        required: true,
        errString:
          "This field is required and must match a valid ID number format",
      },
      {
        name: "name",
        title: "Admin name",
        hint: "Enter the admin's name",
        required: true,
      },
      {
        name: "email",
        title: "Admin email",
        hint: "Enter the admin's email",
        required: true,
        type: "email",
      },
      // TODO validate passwords match
      {
        name: "password",
        title: "Admin password",
        hint: "Password",
        required: true,
        type: "password",
        errString:
          "This field is required and must match the repeat password field",
      },
      {
        name: "repeat_password",
        title: "Repeat password",
        hint: "Repeat password",
        required: true,
        type: "password",
        errString: "This field is required and must match the password field",
      },
      // Election fields
      {
        name: "election_name",
        title: "Election's name",
        hint: "Enter the election's name",
        required: true,
      },
      // TODO start should be before end and before now()
      {
        name: "start",
        title: "Election start",
        hint: "Date and time of election start",
        required: true,
        type: "datetime",
      },
      {
        name: "end",
        title: "Election end",
        hint: "Date and time of election end",
        required: true,
        type: "datetime",
      },
      {
        name: "count_method",
        title: "Vote's count method",
        hint: "Vote count method to use",
        required: true,
        type: "select",
        options: COUNT_METHODS,
      },
      // TODO min should be less or equal to max
      {
        name: "min_candidates",
        title: "Mimimum candidates to vote for",
        hint: "The minimum number of candidates to select",
        required: true,
        type: "number",
        errString:
          "This field is required, must be greater than zero and less or equal to the maximum candidates",
      },
      {
        name: "max_candidates",
        title: "Maximum candidates to vote for",
        hint: "The maximum number of candidates to select",
        required: true,
        type: "number",
        errString:
          "This field is required, must be greater than zero and greater or equal to the minimum candidates",
      },
      // Config fields
      {
        name: "id_formats",
        title: "Allowed identification formats",
        hint: "The allowed identification formats for registration",
        required: true,
        type: "multiselect",
        options: ID_FORMATS,
        errString:
          "This field is required, you should select at least one format",
        validate: validateArrayLengthPositive,
      },
    ],
  };
</script>

<div class="card">
  <div class="card-header">Initialize</div>
  <div class="card-body">
    <Form params={candidateForm} on:executed={siteInitialized} />
  </div>
</div>
