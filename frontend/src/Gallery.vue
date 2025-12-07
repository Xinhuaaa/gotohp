<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { MediaBrowser, ConfigManager, type MediaItem } from '../bindings/app/backend'
import Button from "./components/ui/button/Button.vue"
import MediaItemComponent from './components/MediaItem.vue'
import { toast } from "vue-sonner"

const mediaItems = ref<MediaItem[]>([])
const loading = ref(false)
const pageToken = ref('')
const hasMore = ref(true)
const reachedEnd = ref(false)
const thumbnailSize = ref('medium')
const downloadingItems = ref<Set<string>>(new Set())
const seenMediaKeys = ref<Set<string>>(new Set())

// Load thumbnail size from config
onMounted(async () => {
  try {
    const config = await ConfigManager.GetConfig()
    if (config.thumbnailSize) {
      thumbnailSize.value = config.thumbnailSize
    }
  } catch (error) {
    console.error('Failed to load config:', error)
  }
  loadMediaList()
})

async function loadMediaList() {
  if (loading.value || reachedEnd.value) return
  
  loading.value = true
  try {
    console.log('Loading media list with pageToken:', pageToken.value)
    const result = await MediaBrowser.GetMediaList(pageToken.value, 50)
    console.log('Received result:', result)
    
    if (result && result.items) {
      // Filter out duplicate items based on mediaKey
      const newItems = result.items.filter(item => {
        if (seenMediaKeys.value.has(item.mediaKey)) {
          console.log('Skipping duplicate item:', item.mediaKey)
          return false
        }
        seenMediaKeys.value.add(item.mediaKey)
        return true
      })
      
      console.log(`Adding ${newItems.length} new items (${result.items.length} total in response)`)
      
      // If all items were duplicates or no new items, we've reached the end
      if (newItems.length === 0 && result.items.length > 0) {
        console.log('All items were duplicates - reached end')
        reachedEnd.value = true
        hasMore.value = false
        toast.info('已到底部', {
          description: '没有更多照片了',
        })
      } else {
        mediaItems.value = [...mediaItems.value, ...newItems]
        pageToken.value = result.nextPageToken || ''
        
        // Check if we've reached the end
        if (!result.nextPageToken) {
          console.log('No nextPageToken - reached end')
          reachedEnd.value = true
          hasMore.value = false
          if (newItems.length === 0) {
            toast.info('已到底部', {
              description: '没有更多照片了',
            })
          }
        } else {
          hasMore.value = true
        }
      }
    } else if (!result || !result.items || result.items.length === 0) {
      // No items in response
      console.log('No items in response - reached end')
      reachedEnd.value = true
      hasMore.value = false
      if (mediaItems.value.length > 0) {
        toast.info('已到底部', {
          description: '没有更多照片了',
        })
      }
    }
  } catch (error: any) {
    console.error('Failed to load media list:', error)
    toast.error('Failed to load photos', {
      description: error?.message,
    })
  } finally {
    loading.value = false
  }
}

async function downloadMedia(mediaKey: string, filename: string) {
  if (downloadingItems.value.has(mediaKey)) return
  
  downloadingItems.value.add(mediaKey)
  try {
    const savedPath = await MediaBrowser.DownloadMedia(mediaKey)
    toast.success('Download complete!', {
      description: `Saved to: ${savedPath}`,
    })
  } catch (error: any) {
    console.error('Failed to download media:', error)
    toast.error('Download failed', {
      description: error?.message || 'Unknown error',
    })
  } finally {
    downloadingItems.value.delete(mediaKey)
  }
}

const gridCols = computed(() => {
  switch (thumbnailSize.value) {
    case 'small': return 'grid-cols-6'
    case 'large': return 'grid-cols-2'
    default: return 'grid-cols-4' // medium
  }
})
</script>

<template>
  <div class="w-full h-full flex flex-col p-4 overflow-auto">
    <div class="flex justify-between items-center mb-4">
      <h2 class="text-xl font-semibold">Photo Gallery</h2>
      <div class="flex gap-2">
        <Button 
          v-if="hasMore || loading" 
          variant="outline" 
          @click="loadMediaList"
          :disabled="loading || reachedEnd"
          class="cursor-pointer"
        >
          {{ loading ? 'Loading...' : (reachedEnd ? '已到底部' : 'Load More') }}
        </Button>
        <div v-if="reachedEnd && mediaItems.length > 0" class="text-sm text-muted-foreground flex items-center">
          没有更多照片了
        </div>
      </div>
    </div>

    <div v-if="mediaItems.length === 0 && !loading" class="flex flex-col items-center justify-center h-64 text-muted-foreground">
      <p>No photos found</p>
      <p class="text-sm">Upload some photos to see them here</p>
    </div>

    <div v-if="mediaItems.length > 0" :class="['grid gap-2', gridCols]">
      <div 
        v-for="item in mediaItems" 
        :key="item.mediaKey"
        class="relative group aspect-square bg-secondary rounded overflow-hidden"
      >
        <MediaItemComponent 
          :item="item" 
          :thumbnail-size="thumbnailSize"
          :is-downloading="downloadingItems.has(item.mediaKey)"
          @download="downloadMedia(item.mediaKey, item.filename || 'photo')"
        />
      </div>
    </div>

    <div v-if="loading && mediaItems.length === 0" class="flex items-center justify-center h-64">
      <div class="text-muted-foreground">Loading photos...</div>
    </div>
  </div>
</template>
