<template>
    <nuxt-link :to="'/bbs2'+to" ><slot></slot></nuxt-link>
</template>
<script>
    export default{
        props:{
            to:{
                type:String,
                default:"./",
            },
        }

    }
</script>