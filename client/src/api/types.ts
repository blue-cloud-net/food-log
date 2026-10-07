// 通用类型定义（与后端 model 对应）

export interface Ingredient {
  name: string
  amount: string
  unit: string
}

export interface Step {
  order: number
  content: string
  image?: string
}

export interface Recipe {
  id: string
  user_id: string
  name: string
  description: string
  ingredients: Ingredient[]
  steps: Step[]
  cook_time_minutes: number
  difficulty: '' | 'easy' | 'medium' | 'hard'
  rating: number
  /** 菜谱级标签 id（用户手选） */
  tags: string[]
  /** 食材级标签 id（服务端按规则自动派生，只读） */
  ingredient_tags: string[]
  images: string[]
  is_favorited: boolean
  created_at: string
  updated_at: string
}

export interface Paginated<T> {
  list: T[]
  total: number
  page: number
  page_size: number
}

export interface Restaurant {
  id: string
  user_id: string
  name: string
  address: string
  cuisine_type: string
  description: string
  avg_rating: number
  images: string[]
  lat: number | null
  lng: number | null
  dish_count: number
  created_at: string
  updated_at: string
}

export interface Dish {
  id: string
  restaurant_id: string
  user_id: string
  name: string
  description: string
  price: number | null
  rating: number
  images: string[]
  eaten_at: string | null
  created_at: string
}

export interface RestaurantDetail extends Restaurant {
  dishes: Dish[]
}

export interface UserProfile {
  id: string
  username: string
  email: string
  avatar_url: string
  recipe_count: number
  restaurant_count: number
  dish_count: number
  created_at: string
  updated_at: string
}

export interface Tag {
  id: string
  category_id: string
  /** 为空表示全局预设标签，否则为该用户私有自定义标签 */
  owner_id?: string
  name: string
  /** 同组标签在选择器内互斥（如荤菜/素菜同为 diet） */
  mutex_group?: string
  sort_order: number
  is_system: boolean
}

export interface TagCategory {
  id: string
  /** 为空表示全局预设分类 */
  owner_id?: string
  name: string
  /** el-tag 色型：primary/success/warning/danger/info */
  color: string
  sort_order: number
  is_system: boolean
  tags: Tag[]
}

export interface SearchResult {
  recipes: Recipe[]
  restaurants: Restaurant[]
  dishes: Dish[]
}

export interface Recognition {
  name: string
  ingredients: string[]
  tags: string[]
}

export interface UploadFile {
  url: string
  thumb_url: string
}

export interface ExportData {
  version: string
  exported_at: string
  recipes: Recipe[]
  restaurants: Restaurant[]
  dishes: Dish[]
  favorites: string[]
}
