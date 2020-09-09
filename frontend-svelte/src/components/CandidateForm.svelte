<script>
  import Form from "./Form.svelte";
  import { createEventDispatcher } from "svelte";

  const dispatch = createEventDispatcher();
  let candidateForm = {
    name: "Add new candidate",
    values: "form",
    url: "/api/candidates/add",
    generalError: "Could not add candidate",
    fields: [
      {
        name: "name",
        hint: "Candidate's name",
        required: true,
      },
      {
        name: "presentation",
        hint: "Candidate's presentation",
        type: "textarea",
        required: true,
      },
      {
        name: "image",
        hint: "Upload candidate's image",
        required: true,
        type: "file",
      },
    ],
  };
</script>

<div class="card">
  <div class="card-header">Add candidate</div>
  <div class="card-body">
    <Form
      params={candidateForm}
      on:executed={() => dispatch('executed', true)} />
  </div>
</div>
