<script>
  import Alert from "./Alert.svelte";
  import Loading from "./Loading.svelte";
  import Button from "./Button.svelte";
  import CardTable from "./CardTable.svelte";
  import Form from "./Form.svelte";
  import DownloadFileButton from "./DownloadFileButton.svelte";
  import { get, sortByField } from "../util";

  let files;
  let messages;
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

  getFiles();
  getMessages();

  async function getFiles() {
    files = get("/api/users/files/own", sortByField("name"));
  }
  async function getMessages() {
    messages = get("/api/users/messages/own", sortByField("solved"));
  }

  // TODO handle errors
  async function solveMessage(id) {
    await fetch(`/api/users/messages/solve?id=${id}`);
    getMessages();
  }

  async function deleteFile(id) {
    await fetch(`/api/users/files/delete?id=${id}`);
    getFiles();
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

<CardTable
  headers={['Description', '', '']}
  rows={files}
  error="Could not obtain the uploaded files"
  let:row>
  <td>{row.description}</td>
  <td>
    <DownloadFileButton
      url={`/api/users/files/download?id=${row.id}`}
      filename={row.name} />
  </td>
  <td>
    <Button
      content="Delete file"
      type="danger"
      callback={() => deleteFile(row.id)} />
  </td>
</CardTable>

<div class="card">
  <div class="card-header">Upload file</div>
  <div class="card-body">
    <Form params={uploadFileForm} on:executed={getFiles} />
  </div>
</div>
