<script setup lang="ts"></script>

<template>
  <div>
    <el-container>
      <!-- sidebar -->
      <el-container class="admin-container">
        <!-- header -->
        <el-main>
          <el-scrollbar>
            <router-view />
          </el-scrollbar>
        </el-main>
      </el-container>
    </el-container>
  </div>
</template>

<style scoped lang="scss">
.admin-container {
  flex-direction: column;
  height: 100vh;
}
</style>
