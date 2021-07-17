## 增加文章

/admin/add-article

`request`
```
{
    "tags": ["tag1", "tag2"],
    "data": "This is a test data",
    "custom_article_id": ""
}
```

`response`
```
{
    "errcode": 0,
    "errmsg": "success",
    "article_id": 1,
    "custom_article_id": "zh9mbF6c"
}
```

## 查询文章

/admin/get-article

### 根据数字ID查询
`request`
```
{
    "article_id": 1,
    "custom_article_id": ""
}
```

`response`
```
{
    "errcode": 0,
    "errmsg": "success",
    "id": 1,
    "tag_ids": [
        1,
        2
    ],
    "data": "This is a test data",
    "custom_article_id": "zh9mbF6c",
    "tag_name_array": [
        "tag1",
        "tag2"
    ]
}
```

### 根据字符串ID查询
`request`
```
{
    "article_id": 0,
    "custom_article_id": "zh9mbF6c"
}
```

`response`
```
{
    "errcode": 0,
    "errmsg": "success",
    "id": 1,
    "tag_ids": [
        1,
        2
    ],
    "data": "This is a test data",
    "custom_article_id": "zh9mbF6c",
    "tag_name_array": [
        "tag1",
        "tag2"
    ]
}
```

## 获取指定文章后面的N篇文章

/admin/get-next-articles

### 根据数字ID查询
`request`
```
{
    "article_id": 1,
    "custom_article_id": "",
    "n": 1
}

```

`response`
```
{
    "errcode": 0,
    "errmsg": "success",
    "remote_articles": [
        {
            "id": 2,
            "tag_ids": [
                2,
                3
            ],
            "data": "This is a test data 2.",
            "custom_article_id": "SFEh5uSN",
            "tag_name_array": [
                "tag2",
                "tag3"
            ]
        }
    ]
}

```

### 根据字符串ID查询

`request`
```
{
    "article_id": 0,
    "custom_article_id": "zh9mbF6c",
    "n": 1
}

```

`response`
```
{
    "errcode": 0,
    "errmsg": "success",
    "remote_articles": [
        {
            "id": 2,
            "tag_ids": [
                2,
                3
            ],
            "data": "This is a test data 2.",
            "custom_article_id": "SFEh5uSN",
            "tag_name_array": [
                "tag2",
                "tag3"
            ]
        }
    ]
}

```


## 获取指定文章前面的N篇文章

/admin/get-prev-articles

参考上面



## 获取指定分类下的指定文章后面的N篇文章

/admin/get-next-articles-by-tag

### 以字符串文章ID举例：

`request`
```
{
    "article_id": 0,
    "custom_article_id": "zh9mbF6c",
    "n": 1,
    "tag": "tag2"
}

```

`response`
```
{
    "errcode": 0,
    "errmsg": "success",
    "remote_articles": [
        {
            "id": 2,
            "tag_ids": [
                2,
                3
            ],
            "data": "This is a test data 2.",
            "custom_article_id": "SFEh5uSN",
            "tag_name_array": [
                "tag2",
                "tag3"
            ]
        }
    ]
}
```


## 获取指定分类下的指定文章前面的N篇文章

/admin/get-prev-articles-by-tag

参考前面



## 文章更新

/admin/update-article

`request`
```
{
    "article_id": 1,
    "custom_article_id": "",
    "new_tags": ["tag11", "tag22"],
    "new_data": "new data 1"
}
```

`response`
```
{
    "errcode": 0,
    "errmsg": "success"
}
```


## 删除文章

/admin/delete-article

`request`
```
{
    "article_id": 1,
    "custom_article_id": ""
}
```

`response`
```
{
    "errcode": 0,
    "errmsg": "success"
}
```


## 获取基本信息 

/admin/get-model-info

`response`
```
{
    "errcode": 0,
    "errmsg": "success",
    "article_count": 1,
    "tag_count": 2,
    "max_article_id": 2
}
```

## 获取指定分类下的文章数量

/admin/get-article-count-by-tag

`request`
```
{
    "tag_name": "tag2"
}
```

`response`
```
{
    "errcode": 0,
    "errmsg": "success",
    "article_count": 1
}
```

## 获取指定分类后面的N个分类

/admin/get-next-tags

`request`
```
{
    "tag_name": "tag1",
    "n": 1
}

```

`response`
```
{
    "errcode": 0,
    "errmsg": "success",
    "remote_tags": [
        {
            "id": 2,
            "name": "tag2",
            "article_count": 100
        }
    ]
}

```

## 获取指定分类前面的N个分类

/admin/get-prev-tags

参考前面


## 修改分类名字

/admin/rename-tag

`request`
```
{
    "old_name": "tag2",
    "new_name": "tag299"
}
```

`response`
```
{
    "errcode": 0,
    "errmsg": "success"
}
```




