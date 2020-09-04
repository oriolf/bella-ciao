<script>
  import Loading from "./Loading.svelte";
  import Alert from "./Alert.svelte";
  import Button from "./Buttons/Button.svelte";
  import DownloadFileButton from "./Buttons/DownloadFileButton.svelte";
  import DeleteFileButton from "./Buttons/DeleteFileButton.svelte";
  import { get } from "../util.js";

  export let url;
  export let error;
  let response;
  let page = 1;
  let query = "";

  async function getUsers() {
    response = get(`${url}?page=${page}&items_per_page=10&query=${query}`);
    console.log("USERS:", await response);
  }

  async function validateUser(id) {
    console.log("Validating user...", id);
  }

  getUsers();
</script>

<style>
  .user-container {
    border: 1px solid rgba(0, 0, 0, 0.125);
    border-radius: 3px;
    margin-bottom: 5px;
  }
</style>


{#await response}
  <Loading />
{:then resp}
  <div class="card">
    <div class="card-body">
      {#each resp.users || [] as user}
        <div class="container user-container">
          <div class="row" style="margin-top: 5px;">
            <div class="col-6">
              <h6>{user.unique_id}</h6>
              <p>{user.name} <em>({user.email})</em></p>
            </div>
            <div class="col-3">
              <Button content="Add message" callback={() => {}} />
            </div>
            <div class="col-3">
              <Button
                content="Validate"
                type="success"
                callback={() => validateUser(user.id)} />
            </div>
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
                <DeleteFileButton id={file.id} on:executed={getUsers} />
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
    </div>
  </div>
{:catch _}
  <Alert type="danger" content={error} />
{/await}
