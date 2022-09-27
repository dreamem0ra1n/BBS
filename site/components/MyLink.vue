<template>
    <a :href="to" target="_blank"><slot></slot></a>
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