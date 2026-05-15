package taxonomy

import (
	"time"

	"github.com/google/uuid"
)

type CategoryConverter struct{}

func (c *CategoryConverter) CreateDTOToModel(dto CategoryCreateDTO) Category {
	return Category{
		ID:          uuid.New(),
		Name:        dto.Name,
		Slug:        dto.Slug,
		Description: dto.Description,
		CreatedAt:   time.Now(),
	}
}

func (c *CategoryConverter) UpdateDTOToModel(dto CategoryUpdateDTO) Category {
	return Category{
		Name:        dto.Name,
		Slug:        dto.Slug,
		Description: dto.Description,
	}
}

func (c *CategoryConverter) ModelToResponseDTO(model Category) CategoryResponseDTO {
	return CategoryResponseDTO(model)
}

func (c *CategoryConverter) ModelsToResponseDTOs(models []Category) []CategoryResponseDTO {
	dtos := make([]CategoryResponseDTO, len(models))
	for i, model := range models {
		dtos[i] = c.ModelToResponseDTO(model)
	}
	return dtos
}

type TagConverter struct{}

func (c *TagConverter) CreateDTOToModel(dto TagCreateDTO) Tag {
	return Tag{
		ID:        uuid.New(),
		Name:      dto.Name,
		Slug:      dto.Slug,
		CreatedAt: time.Now(),
	}
}

func (c *TagConverter) UpdateDTOToModel(dto TagUpdateDTO) Tag {
	return Tag{
		Name: dto.Name,
		Slug: dto.Slug,
	}
}

func (c *TagConverter) ModelToResponseDTO(model Tag) TagResponseDTO {
	return TagResponseDTO(model)
}

func (c *TagConverter) ModelsToResponseDTOs(models []Tag) []TagResponseDTO {
	dtos := make([]TagResponseDTO, len(models))
	for i, model := range models {
		dtos[i] = c.ModelToResponseDTO(model)
	}
	return dtos
}
