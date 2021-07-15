from django.db import models
from django.utils import timezone

class Tag(models.Model):
    tag_id = models.BigAutoField(primary_key=True)
    tag = models.CharField(max_length=50)

class Author(models.Model):
    author_id = models.BigAutoField(primary_key=True)
    name = models.TextField()

class Book(models.Model):
    FICTION = 1
    NONFICTION = 2
    BOOK_TYPE_CHOICES = [
        (FICTION, 'FICTION'),
        (NONFICTION, 'NONFICTION'),
    ]
    book_id = models.BigAutoField(primary_key=True)
    author = models.ForeignKey(Author, on_delete=models.CASCADE)
    isbn = models.CharField(max_length=255)
    book_type = models.IntegerField(choices = BOOK_TYPE_CHOICES)
    title = models.CharField(max_length=255)
    year = models.IntegerField(default=2000)
    available = models.DateTimeField(default=timezone.now)
    tags = models.ManyToManyField(Tag)
