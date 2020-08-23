<script>
  import { createEventDispatcher } from "svelte";
  import {
    submitForm,
    paramsFunc,
    validationFuncs,
    validateNonEmpty,
    formatValidationError,
  } from "../util.js";

  // TODO validate all fields from login and register
  // TODO show errors on the appropriate field, not on the bottom
  const dispatch = createEventDispatcher();
  let showError = false;
  let loggingIn = true;

  async function login(e) {
    const res = await submitForm(e, "/api/auth/login");
    if (res.ok) {
      dispatch("loggedin", true);
    } else {
      showError = true;
    }
  }

  const setRegister = () => {
    loggingIn = false;
  };
  const setLogin = () => {
    loggingIn = true;
  };
  let registerValFunc = validationFuncs([validateNonEmpty("name")]);
  let registerParamsFunc = paramsFunc(registerValFunc);
  let registerValidationErrors = [];

  async function register(e) {
    const res = await submitForm(e, "/api/auth/register", registerParamsFunc);
    if (res.ok) {
      loggingIn = true;
    } else if (res.validationErrors) {
      registerValidationErrors = res.validationErrors;
    } else {
      showError = true;
    }
  }
</script>

{#if loggingIn}
  <div class="card">
    <div class="card-header">Log in</div>
    <div class="card-body">
      <form on:submit={login}>
        <div class="form-group">
          <input
            class="form-control"
            id="unique_id"
            name="unique_id"
            aria-describedby="uniqueIdentifierHelp"
            placeholder="Enter user (identification number: DNI, NIE, etc.)" />
        </div>
        <div class="form-group">
          <input
            type="password"
            class="form-control"
            id="password"
            name="password"
            placeholder="Password" />
        </div>
        <button type="submit" class="btn btn-primary">Log in</button>
        <a href="." class="btn" on:click={setRegister}>Register</a>
      </form>
    </div>
    {#if showError}
      <div class="card-footer">
        <div class="alert alert-danger" role="alert" style="margin-bottom: 0;">
          The user or password is invalid
        </div>
      </div>
    {/if}
  </div>
{:else}
  <div class="card">
    <div class="card-header">Register</div>
    <div class="card-body">
      <form on:submit={register}>
        <div class="form-group">
          <input
            class="form-control"
            id="unique_id"
            name="unique_id"
            aria-describedby="uniqueIdentifierHelp"
            placeholder="Enter user (identification number: DNI, NIE, etc.)" />
        </div>
        <div class="form-group">
          <input
            class="form-control"
            id="name"
            name="name"
            aria-describedby="nameHelp"
            placeholder="Enter your name" />
        </div>
        <div class="form-group">
          <input
            type="email"
            class="form-control"
            id="email"
            name="email"
            aria-describedby="emailHelp"
            placeholder="Enter your email" />
        </div>
        <div class="form-group">
          <input
            type="password"
            class="form-control"
            id="password"
            name="password"
            placeholder="Password" />
        </div>
        <div class="form-group">
          <input
            type="password"
            class="form-control"
            id="re_password"
            name="re_password"
            placeholder="Repeat password" />
        </div>
        <button type="submit" class="btn btn-primary">Register</button>
        <a href="." class="btn" on:click={setLogin}>Log in</a>
      </form>
    </div>
    {#if showError || registerValidationErrors.length > 0}
      <div class="card-footer">
        <div class="alert alert-danger" role="alert" style="margin-bottom: 0;">
          {#if showError}
            The user is not a valid identification number (DNI, NIE...), the
            name or email are empty, or the passwords don't match
          {:else}
            {#each registerValidationErrors as err}
              {formatValidationError(err)}
            {/each}
          {/if}
        </div>
      </div>
    {/if}
  </div>
{/if}
