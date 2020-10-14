<script lang="ts">
  import Form from "./Form.svelte";
  import { createEventDispatcher } from "svelte";
  import type { FormParams } from "../types/models.type";
  import { _ } from "svelte-i18n";

  const dispatch = createEventDispatcher();
  let candidateForm: FormParams = {
    name: "comp.candidate_form.form_name",
    values: "form",
    url: "/api/candidates/add",
    generalError: "comp.candidate_form.general_error",
    fields: [
      {
        name: "name",
        hint: "comp.candidate_form.name",
        required: true,
      },
      {
        name: "presentation",
        hint: "comp.candidate_form.presentation",
        type: "textarea",
        required: true,
      },
      {
        name: "image",
        hint: "comp.candidate_form.image",
        required: true,
        type: "file",
      },
    ],
  };
</script>

<div class="card">
  <div class="card-header">{$_("comp.candidate_form.add_candidate")}</div>
  <div class="card-body">
    <Form
      params={candidateForm}
      on:executed={() => dispatch('executed', true)} />
  </div>
</div>
