<script>
  import Error from "../routes/_error.svelte";
  import LoginForm from "./LoginForm.svelte";
  import RoleNone from "./RoleNone.svelte";
  import RoleValidated from "./RoleValidated.svelte";
  import RoleAdmin from "./RoleAdmin.svelte";

  async function whoami() {
    let res = await fetch("/users/whoami");
    if (!res.ok) {
      throw new Error("Not logged in");
    }
    return res.json();
  }
  let promise = whoami(); // TODO should be called in onMount
  function getUser() {
    promise = whoami();
  }
</script>

<svelte:head>
  <title>Bella Ciao</title>
</svelte:head>

{#await promise}
  <p>...waiting</p>
{:then user}
  {#if user.role === 'none'}
    <RoleNone />
  {:else if user.role === 'validated'}
    <RoleValidated />
  {:else if user.role === 'admin'}
    <RoleAdmin />
  {:else}
    <p>Unknown user role {user.role}</p>
  {/if}
{:catch _}
  <LoginForm on:loggedin={getUser} />
{/await}
