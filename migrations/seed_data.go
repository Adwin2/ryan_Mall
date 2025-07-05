package main

import (
	"log"
	"ryan-mall/internal/config"
	"ryan-mall/internal/model"
	"ryan-mall/pkg/database"
)

// 这个文件用于初始化测试数据
// 在开发阶段可以运行这个脚本来创建一些基础数据

func main() {
	// 1. 加载配置
	cfg := config.LoadConfig()
	
	// 2. 初始化数据库连接
	if err := database.InitMySQL(cfg); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer database.Close()
	
	// 3. 获取数据库实例
	db := database.GetDB()
	
	// 4. 创建测试分类
	categories := []model.Category{
		{Name: "电子产品", ParentID: 0, SortOrder: 1},
		{Name: "服装鞋帽", ParentID: 0, SortOrder: 2},
		{Name: "图书音像", ParentID: 0, SortOrder: 3},
	}
	
	for _, category := range categories {
		// 检查分类是否已存在
		var existingCategory model.Category
		if err := db.Where("name = ?", category.Name).First(&existingCategory).Error; err != nil {
			// 分类不存在，创建新分类
			if err := db.Create(&category).Error; err != nil {
				log.Printf("Failed to create category %s: %v", category.Name, err)
			} else {
				log.Printf("Created category: %s", category.Name)
			}
		} else {
			log.Printf("Category already exists: %s", category.Name)
		}
	}
	
	// 5. 创建子分类
	var electronicsCategory model.Category
	db.Where("name = ?", "电子产品").First(&electronicsCategory)
	
	var clothingCategory model.Category
	db.Where("name = ?", "服装鞋帽").First(&clothingCategory)
	
	subCategories := []model.Category{
		{Name: "手机数码", ParentID: electronicsCategory.ID, SortOrder: 1},
		{Name: "电脑办公", ParentID: electronicsCategory.ID, SortOrder: 2},
		{Name: "男装", ParentID: clothingCategory.ID, SortOrder: 1},
		{Name: "女装", ParentID: clothingCategory.ID, SortOrder: 2},
	}
	
	for _, subCategory := range subCategories {
		var existingSubCategory model.Category
		if err := db.Where("name = ? AND parent_id = ?", subCategory.Name, subCategory.ParentID).First(&existingSubCategory).Error; err != nil {
			if err := db.Create(&subCategory).Error; err != nil {
				log.Printf("Failed to create sub-category %s: %v", subCategory.Name, err)
			} else {
				log.Printf("Created sub-category: %s", subCategory.Name)
			}
		} else {
			log.Printf("Sub-category already exists: %s", subCategory.Name)
		}
	}
	
	// 6. 创建测试商品
	var phoneCategory model.Category
	db.Where("name = ?", "手机数码").First(&phoneCategory)
	
	var computerCategory model.Category
	db.Where("name = ?", "电脑办公").First(&computerCategory)
	
	var menClothingCategory model.Category
	db.Where("name = ?", "男装").First(&menClothingCategory)
	
	originalPrice1 := 8999.00
	originalPrice2 := 13999.00
	originalPrice3 := 799.00
	
	products := []model.Product{
		{
			Name:          "iPhone 15 Pro",
			Description:   stringPtr("苹果最新款手机，性能强劲，拍照效果出色"),
			CategoryID:    phoneCategory.ID,
			Price:         7999.00,
			OriginalPrice: &originalPrice1,
			Stock:         100,
			MainImage:     stringPtr("https://example.com/iphone15.jpg"),
			Images:        model.JSONArray{"https://example.com/iphone15-1.jpg", "https://example.com/iphone15-2.jpg"},
		},
		{
			Name:          "MacBook Pro",
			Description:   stringPtr("苹果笔记本电脑，适合开发和设计工作"),
			CategoryID:    computerCategory.ID,
			Price:         12999.00,
			OriginalPrice: &originalPrice2,
			Stock:         50,
			MainImage:     stringPtr("https://example.com/macbook.jpg"),
			Images:        model.JSONArray{"https://example.com/macbook-1.jpg", "https://example.com/macbook-2.jpg"},
		},
		{
			Name:          "Nike运动鞋",
			Description:   stringPtr("舒适透气的运动鞋，适合日常运动"),
			CategoryID:    menClothingCategory.ID,
			Price:         599.00,
			OriginalPrice: &originalPrice3,
			Stock:         200,
			MainImage:     stringPtr("https://example.com/nike.jpg"),
			Images:        model.JSONArray{"https://example.com/nike-1.jpg", "https://example.com/nike-2.jpg"},
		},
	}
	
	for _, product := range products {
		var existingProduct model.Product
		if err := db.Where("name = ?", product.Name).First(&existingProduct).Error; err != nil {
			if err := db.Create(&product).Error; err != nil {
				log.Printf("Failed to create product %s: %v", product.Name, err)
			} else {
				log.Printf("Created product: %s", product.Name)
			}
		} else {
			log.Printf("Product already exists: %s", product.Name)
		}
	}
	
	log.Println("Seed data initialization completed!")
}

// stringPtr 返回字符串指针
func stringPtr(s string) *string {
	return &s
}
