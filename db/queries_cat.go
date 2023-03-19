package db

import "git.sr.ht/~bouncepaw/betula/types"

func CategoriesForPost(id int) (cats []types.Category) {
	q := `
select distinct CatName
from CategoriesToPosts
where PostID = ?;
`
	rows := mustQuery(q, id)
	for rows.Next() {
		var cat types.Category
		mustScan(rows, &cat.Name)
		cats = append(cats, cat)
	}
	return cats
}

func DescriptionForCategory(catName string) (myco string) {
	const q = `
select Description from CategoryDescriptions where CatName = ?;
`
	rows := mustQuery(q, catName)
	for rows.Next() { // 0 or 1
		mustScan(rows, &myco)
		break
	}
	_ = rows.Close()

	return myco
}

// Categories returns all categories found on posts one has access to. They all have PostCount set to a non-zero value.
func Categories(authorized bool) (cats []types.Category) {
	q := `
select
   CatName, 
   count(PostID)
from
   CategoriesToPosts
inner join 
    (select ID from main.Posts where DeletionTime is null and (Visibility = 1 or ?)) 
as 
	Filtered
on 
    CategoriesToPosts.PostID = Filtered.ID
group by
	CatName;
`
	rows := mustQuery(q, authorized)
	for rows.Next() {
		var cat types.Category
		mustScan(rows, &cat.Name, &cat.PostCount)
		cats = append(cats, cat)
	}
	return cats
}

func SetCategoriesFor(postID int, categories []types.Category) {
	const qDelete = `delete from CategoriesToPosts where PostID = ?;`
	mustExec(qDelete, postID)

	var qAdd = `insert into CategoriesToPosts (CatName, PostID) values (?, ?);`
	for _, cat := range categories {
		if cat.Name == "" {
			continue
		}
		mustExec(qAdd, cat.Name, postID)
	}
}

func EditCategory(category types.Category, newName, newDescription string) {
	// FIXME: Research transactions and probably use them here.
	const q = `
update CategoriesToPosts
set
    CatName = ?
where
    CatName = ?;

replace into CategoryDescriptions (CatName, Description)
values (?, ?);
`
	mustExec(q, newName, category.Name, newName, newDescription)
}

func CategoryExists(category types.Category) (has bool) {
	const q = `select exists(select 1 from CategoriesToPosts where CatName = ?);`
	rows := mustQuery(q, category.Name)
	rows.Next()
	mustScan(rows, &has)
	_ = rows.Close()
	return has
}
