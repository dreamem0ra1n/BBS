<template>
  <user-profile-layout>
    <nuxt-child />
  </user-profile-layout>
</template>

<script>
export default {
  middleware: 'authenticated',
}
</script>
