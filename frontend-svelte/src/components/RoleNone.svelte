<script>
  import Alert from "./Alert.svelte";
  import Button from "./Buttons/Button.svelte";
  import CardTable from "./CardTable.svelte";
  import Form from "./Form.svelte";
  import UserFiles from "./UserFiles.svelte";
  import { get, sortByField } from "../util";

  let messages;
  let reloadFiles = 0;
  let uploadFileForm = {
    name: "Upload file",
    values: "form",
    url: "/api/users/files/upload",
    generalError: "Could not upload file",
    fields: [
      {
        name: "description",
        hint: "Description of the file contents",
        required: true,
      },
      {
        name: "file",
        hint: "Upload file",
        required: true,
        type: "file",
      },
    ],
  };

  getMessages();

  async function getMessages() {
    messages = get("/api/users/messages/own", sortByField("solved"));
  }

  // TODO handle errors
  async function solveMessage(id) {
    await fetch(`/api/users/messages/solve?id=${id}`);
    getMessages();
  }
</script>

<Alert
  content="You still have not been validated. Please review and solve the
  validators' messages and upload the required files" />

<h2>Messages from validation</h2>
<p>
  Please review these messages, perform the requested action, and mark as solved
  when done.
</p>

<CardTable
  headers={['Message', '', '']}
  rows={messages}
  let:row
  error="Could not obtain the validation messages">
  <td>{row.content}</td>
  <td>
    {#if row.solved}
      <em>(already solved)</em>
    {:else}
      <Button content="Mark as solved" callback={() => solveMessage(row.id)} />
    {/if}
  </td>
</CardTable>

<h2>Uploaded files</h2>
<p>Upload the required files and remove those that are no longer necessary.</p>

<UserFiles {reloadFiles} />

<div class="card">
  <div class="card-header">Upload file</div>
  <div class="card-body">
    <Form params={uploadFileForm} on:executed={() => (reloadFiles += 1)} />
  </div>
</div>
