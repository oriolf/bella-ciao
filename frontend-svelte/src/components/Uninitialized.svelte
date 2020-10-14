<script lang="ts">
  import Form from "./Form.svelte";
  import { COUNT_METHODS, ID_FORMATS } from "../constants.js";
  import { validateArrayLengthPositive } from "../util";
  import type { StringMap, FormParams, JsonValue } from "../types/models.type";
  import { _ } from "svelte-i18n";

  function siteInitialized() {
    window.location.href = "/";
  }

  function composeTime(v: StringMap, field: string) {
    return new Date(
      v[field + "_date"] + " " + v[field + "_time"]
    ).toISOString();
  }

  function groupParameters(v: StringMap): JsonValue {
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

  let candidateForm: FormParams = {
    name: "comp.uninitialized.initialize",
    url: "/api/initialize",
    generalError: "comp.uninitialized.initialize_err",
    jsonFunc: groupParameters,
    fields: [
      // Admin fields
      {
        name: "unique_id",
        title: "comp.uninitialized.unique_id_title",
        hint: "comp.uninitialized.unique_id_hint",
        required: true,
        errString: "comp.uninitialized.unique_id_err",
      },
      {
        name: "name",
        title: "comp.uninitialized.name_title",
        hint: "comp.uninitialized.name_hint",
        required: true,
      },
      {
        name: "email",
        title: "comp.uninitialized.email_title",
        hint: "comp.uninitialized.email_hint",
        required: true,
        type: "email",
      },
      // TODO validate passwords match
      {
        name: "password",
        title: "comp.uninitialized.password_title",
        hint: "comp.uninitialized.password_hint",
        required: true,
        type: "password",
        errString: "comp.uninitialized.password_err",
      },
      {
        name: "repeat_password",
        title: "comp.uninitialized.rpassword_title",
        hint: "comp.uninitialized.rpassword_hint",
        required: true,
        type: "password",
        errString: "comp.uninitialized.password_err",
      },
      // Election fields
      {
        name: "election_name",
        title: "comp.uninitialized.election_name_title",
        hint: "comp.uninitialized.election_name_hint",
        required: true,
      },
      // TODO start should be before end and before now()
      {
        name: "start",
        title: "comp.uninitialized.start_title",
        hint: "comp.uninitialized.start_hint",
        required: true,
        type: "datetime",
      },
      {
        name: "end",
        title: "comp.uninitialized.end_title",
        hint: "comp.uninitialized.end_hint",
        required: true,
        type: "datetime",
      },
      {
        name: "count_method",
        title: "comp.uninitialized.count_method_title",
        hint: "comp.uninitialized.count_method_hint",
        required: true,
        type: "select",
        options: COUNT_METHODS,
      },
      // TODO min should be less or equal to max
      {
        name: "min_candidates",
        title: "comp.uninitialized.min_candidates_title",
        hint: "comp.uninitialized.min_candidates_hint",
        required: true,
        type: "number",
        errString: "comp.uninitialized.min_candidates_err",
      },
      {
        name: "max_candidates",
        title: "comp.uninitialized.max_candidates_title",
        hint: "comp.uninitialized.max_candidates_hint",
        required: true,
        type: "number",
        errString: "comp.uninitialized.max_candidates_err",
      },
      // Config fields
      {
        name: "id_formats",
        title: "comp.uninitialized.id_formats_title",
        hint: "comp.uninitialized.id_formats_hint",
        required: true,
        type: "multiselect",
        options: ID_FORMATS,
        errString: "comp.uninitialized.id_formats_err",
        validate: validateArrayLengthPositive,
      },
    ],
  };
</script>

<div class="card">
  <div class="card-header">{$_("comp.uninitialized.initialize")}</div>
  <div class="card-body">
    <Form params={candidateForm} on:executed={siteInitialized} />
  </div>
</div>
