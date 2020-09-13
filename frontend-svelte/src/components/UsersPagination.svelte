<script>
  import Loading from "./Loading.svelte";
  import Alert from "./Alert.svelte";
  import Pagination from "./Pagination.svelte";
  import Button from "./Buttons/Button.svelte";
  import DownloadFileButton from "./Buttons/DownloadFileButton.svelte";
  import DeleteFileButton from "./Buttons/DeleteFileButton.svelte";
  import { get, submitFormJSON } from "../util.ts";

  export let url;
  export let error;
  export let unvalidated;
  let response;
  let page = 1;
  let itemsPerPage = 10;
  let query = "";
  let timer;
  let userMessageID;
  let userMessage = "";

  const debounce = (v) => {
    clearTimeout(timer);
    timer = setTimeout(() => {
      query = v;
    }, 400);
  };

  $: getUsers(page, query);

  async function getUsers(pg, qry) {
    response = get(
      `${url}?page=${pg}&items_per_page=${itemsPerPage}&query=${qry}`
    );
  }

  async function validateUser(id) {
    await fetch(`/api/users/validate?id=${id}`);
    getUsers(page, query);
  }

  function beginAddMessage(userID) {
    userMessageID = userID;
    jQuery("#addMessageModal").on("shown.bs.modal", function () {
      jQuery("#userMessageInput").trigger("focus");
    });
  }

  async function addMessage() {
    await submitFormJSON("/api/users/messages/add", {
      user_id: userMessageID,
      content: userMessage,
    });
    jQuery("#addMessageModal").modal("hide");
    userMessage = "";
    getUsers(page, query);
  }
</script>

<style>
  .user-container {
    border: 1px solid rgba(0, 0, 0, 0.125);
    border-radius: 3px;
    margin-bottom: 5px;
  }
</style>

<div class="card">
  <div class="card-body">
    <div class="input-group mb-3">
      <input
        type="text"
        class="form-control"
        placeholder="Filter by user identifier"
        aria-label="Filter by user identifier"
        aria-describedby="filter-users"
        on:keyup={({ target: { value } }) => debounce(value)} />
    </div>
    {#await response}
      <Loading />
    {:then resp}
      {#each resp.users || [] as user}
        <div class="container user-container">
          <div class="row" style="margin-top: 5px;">
            <div class="col-6">
              <h6>
                {user.unique_id}
                {#if user.role === 'admin'}<em>(admin)</em>{/if}
              </h6>
              <p>{user.name} <em>({user.email})</em></p>
            </div>
            {#if unvalidated}
              <div class="col-3">
                <button
                  type="button"
                  class="align-middle btn btn-sm btn-outline-primary"
                  style="width: 100%;"
                  data-toggle="modal"
                  data-target="#addMessageModal"
                  on:click={() => beginAddMessage(user.id)}>
                  Add message
                </button>
              </div>
              <div class="col-3">
                <Button
                  content="Validate"
                  type="success"
                  callback={() => validateUser(user.id)} />
              </div>
            {/if}
          </div>

          <div class="row">
            <div class="col-12">
              <h6>Files</h6>
            </div>
          </div>

          {#each user.files || [] as file}
            <div class="row" style="margin-top: 2px;">
              <div class="col-6" style="padding-left: 32px;">
                {file.description}
              </div>
              <div class="col-3">
                <DownloadFileButton id={file.id} filename={file.name} />
              </div>
              <div class="col-3">
                <DeleteFileButton
                  id={file.id}
                  on:executed={() => getUsers(page, query)} />
              </div>
            </div>
          {/each}

          <div class="row">
            <div class="col-12">
              <h6>Messages</h6>
            </div>
          </div>

          {#each user.messages || [] as message}
            <div class="row" style="margin: 2px 0 0 5px;">
              <div class="col-12">
                {#if message.solved}
                  <span class="badge badge-primary">Already solved</span>
                {:else}
                  <span class="badge badge-warning">Not solved yet</span>
                {/if}
                {message.content}
              </div>
            </div>
          {/each}
        </div>
      {/each}
      <Pagination bind:page {itemsPerPage} totalItems={resp.total} />
    {:catch _}
      <Alert type="danger" content={error} />
    {/await}
  </div>
</div>

<!-- Modal -->
<div
  class="modal fade"
  id="addMessageModal"
  tabindex="-1"
  role="dialog"
  aria-labelledby="addMessageModalLabel"
  aria-hidden="true">
  <div class="modal-dialog" role="document">
    <div class="modal-content">
      <div class="modal-header">
        <h5 class="modal-title" id="addMessageModalLabel">Add message</h5>
      </div>
      <div class="modal-body">
        <div class="form-group">
          <textarea
            id="userMessageInput"
            class="form-control"
            rows="3"
            bind:value={userMessage} />
        </div>
      </div>
      <div class="modal-footer">
        <button
          type="button"
          class="btn btn-secondary"
          data-dismiss="modal">Close</button>
        <button type="button" class="btn btn-primary" on:click={addMessage}>
          Add message
        </button>
      </div>
    </div>
  </div>
</div>
