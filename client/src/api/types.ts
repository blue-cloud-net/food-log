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
  tags: string[]
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

export interface TagCategory {
  name: string
  tags: string[]
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
