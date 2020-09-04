<script>
  import CardPagination from "./CardPagination.svelte";
  import Button from "./Buttons/Button.svelte";
  import Alert from "./Alert.svelte";
  import DownloadFileButton from "./Buttons/DownloadFileButton.svelte";
  import DeleteFileButton from "./Buttons/DeleteFileButton.svelte";
  import { get } from "../util.js";

  let validated;
  let unvalidated;

  async function getValidated(page, query) {
    validated = get("/api/users/validated/get");
  }
  async function getUnvalidated(page, query) {
    unvalidated = get("/api/users/unvalidated/get");
    console.log("USERS:", await unvalidated);
  }
  async function validateUser(id) {
    console.log("Validating user...", id);
  }

  getValidated();
  getUnvalidated();
</script>

<style>
  .user-container {
    border: 1px solid rgba(0, 0, 0, 0.125);
    border-radius: 3px;
    margin-bottom: 5px;
  }
</style>

<h2>Users pending validation</h2>

<CardPagination
  rows={unvalidated}
  error="Could not get users pending validation"
  let:row>
  <div class="container user-container">
    <div class="row" style="margin-top: 5px;">
      <div class="col-6">
        <h6>{row.unique_id}</h6>
        <p>{row.name} <em>({row.email})</em></p>
      </div>
      <div class="col-3">
        <Button content="Add message" callback={() => {}} />
      </div>
      <div class="col-3">
        <Button
          content="Validate"
          type="success"
          callback={() => validateUser(row.id)} />
      </div>
    </div>

    <div class="row">
      <div class="col-12">
        <h6>Files</h6>
      </div>
    </div>

    {#each row.files || [] as file}
      <div class="row" style="margin-top: 2px;">
        <div class="col-6" style="padding-left: 32px;">{file.description}</div>
        <div class="col-3">
          <DownloadFileButton id={file.id} filename={file.name} />
        </div>
        <div class="col-3">
          <DeleteFileButton id={file.id} on:executed={getUnvalidated} />
        </div>
      </div>
    {/each}

    <div class="row">
      <div class="col-12">
        <h6>Messages</h6>
      </div>
    </div>

    {#each row.messages || [] as message}
      <div class="row" style="margin: 2px 0 0 5px;">
        <div class="col-12">
          {#if message.solved}
            <span class="badge badge-primary">Already solved</span>
          {:else}<span class="badge badge-warning">Not solved yet</span>{/if}
          {message.content}
        </div>
      </div>
    {/each}
  </div>
</CardPagination>

<h2>Validated users</h2>

<CardPagination rows={validated} error="Could not get validated users" let:row>
  <p>{row.unique_id}</p>
</CardPagination>
