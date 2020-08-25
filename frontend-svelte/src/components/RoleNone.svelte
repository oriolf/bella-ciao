<script>
  import Alert from "./Alert.svelte";
  import Loading from "./Loading.svelte";
  import Button from "./Button.svelte";
  import DownloadFileButton from "./DownloadFileButton.svelte";
  import { sortByField } from "../util";

  let files;
  let messages;

  getFiles();
  getMessages();

  async function getFiles() {
    files = get("/api/users/files/own", sortByField("name"));
  }
  async function getMessages() {
    messages = get("/api/users/messages/own", sortByField("solved"));
  }
  async function get(url, sortFunc) {
    let res = await fetch(url);
    if (!res.ok) {
      throw new Error("Could not get files");
    }
    return (await res.json()).sort(sortFunc);
  }

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
{#await messages}
  <Loading />
{:then msgs}
  <table class="table">
    <thead>
      <tr>
        <th>Message</th>
        <th />
      </tr>
    </thead>
    <tbody>
      {#each msgs as msg}
        <tr>
          <td>{msg.content}</td>
          <td>
            {#if msg.solved}
              <em>(already solved)</em>
            {:else}
              <Button
                content="Mark as solved"
                callback={() => solveMessage(msg.id)} />
            {/if}
          </td>
        </tr>
      {/each}
    </tbody>
  </table>
{:catch _}
  <Alert type="danger" content="Could not obtain the validation messages" />
{/await}

<h2>Uploaded files</h2>
<p>Upload the required files and remove those that are no longer necessary.</p>

{#await files}
  <Loading />
{:then fls}
  {#if fls.length > 0}
    <table class="table">
      <thead>
        <tr>
          <th>Description</th>
          <th />
          <th />
        </tr>
      </thead>
      <tbody>
        {#each fls as file}
          <tr>
            <td>{file.description}</td>
            <td>
              <DownloadFileButton
                url={`/api/users/files/download?id=${file.id}`}
                filename={file.name} />
            </td>
            <td>
              <Button
                content="Delete file"
                type="danger"
                callback={() => deleteFile(file.id)} />
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
  {/if}
{:catch _}
  <Alert type="danger" content="Could not obtain the uploaded files" />
{/await}
