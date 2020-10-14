<script lang="ts">
  // TODO change $_ for $t, wrap it with util func such that the comp.component_name prefix does not need to be repeated
  import LoginForm from "../components/LoginForm.svelte";
  import RoleNone from "../components/RoleNone.svelte";
  import RoleValidated from "../components/RoleValidated.svelte";
  import RoleAdmin from "../components/RoleAdmin.svelte";
  import Loading from "../components/Loading.svelte";
  import { user } from "../store";
  import { whoami } from "../util";

  let promise = whoami(user);
  
  function getUser() {
    promise = whoami(user);
  }
</script>

<svelte:head>
  <title>Bella Ciao</title>
</svelte:head>

{#await promise}
  <Loading />
{:then _}
  {#if $user.role === 'none'}
    <RoleNone />
  {:else if $user.role === 'validated'}
    <RoleValidated />
  {:else if $user.role === 'admin'}
    <RoleAdmin />
  {:else}
    <p>Unknown user role {$user.role}</p>
  {/if}
{:catch _}
  <LoginForm on:loggedin={getUser} />
{/await}
