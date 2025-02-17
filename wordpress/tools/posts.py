import urllib.parse
import os
from tools.helper import WORDPRESS_API_URL, tool_registry, str_to_bool, is_valid_iso8601


def _format_posts_response(response_json: dict):
    new_response_json = {}
    keys = [
        "id",
        "date",
        "date_gmt",
        "guid",
        "modified",
        "modified_gmt",
        "slug",
        "status",
        "type",
        "link",
        "title",
        "content",
        "excerpt",
        "author",
    ]
    for key in keys:
        new_response_json[key] = response_json[key]
    return new_response_json


@tool_registry.register("RetrievePost")
def retrieve_post(client):
    post_id = os.environ["POST_ID"]
    url = f"{WORDPRESS_API_URL}/posts/{post_id}"

    context = os.getenv("CONTEXT", "view").lower()
    context_enum = {"view", "embed", "edit"}
    if context not in context_enum:
        raise ValueError(
            f"Error: Invalid context: {context}. context must be one of: {context_enum}."
        )
    url += f"?context={context}"
    password = os.getenv("PASSWORD", None)
    if password:
        url += f"&password={password}"

    response = client.get(url)
    if response.status_code == 200:
        return response.json()
    else:
        print(
            f"Failed to retrieve post. Error: {response.status_code}, {response.text}"
        )


@tool_registry.register("ListPosts")
def list_posts(client):
    url = f"{WORDPRESS_API_URL}/posts/"

    suffix = ""
    context = os.getenv("CONTEXT", "view").lower()
    context_enum = {"view", "embed", "edit"}
    if context not in context_enum:
        raise ValueError(
            f"Invalid context. Valid context must be one of: {context_enum}"
        )
    suffix += f"?context={context}"
    page = os.getenv("PAGE", 1)
    suffix += f"&page={page}"
    per_page = os.getenv("PER_PAGE", 10)
    suffix += f"&per_page={per_page}"
    author_ids = os.getenv("AUTHOR_IDS", None)  # a list of comma separated author ids
    if author_ids:
        for author_id in author_ids.split(","):
            if not author_id.isdigit():
                raise ValueError(
                    f"Error: Invalid author_ids: {author_id}. Author_ids must be a list of integers separated by commas."
                )
        suffix += f"&author={author_ids}"
    search_query = os.getenv("SEARCH_QUERY", None)
    if search_query:
        suffix += f"&search={search_query}"
    statuses = os.getenv(
        "STATUSES", "publish"
    ).lower()  # a list of comma separated statuses
    status_enum = [
        "publish",
        "future",
        "draft",
        "pending",
        "private",
        "trash",
        "auto-draft",
        "inherit",
        "request-pending",
        "request-confirmed",
        "request-failed",
        "request-completed",
    ]
    for s in statuses.split(","):
        if s not in status_enum:
            raise ValueError(
                f"Error: Invalid status: {s}. status must be a comma separated list of valid statuses: {status_enum}."
            )
    suffix += f"&status={statuses}"

    publish_after = os.getenv("PUBLISH_AFTER", None)
    if publish_after:
        if not is_valid_iso8601(publish_after):
            raise ValueError(
                f"Error: Invalid publish_after: {publish_after}. publish_after must be a valid ISO 8601 date string, in the format of YYYY-MM-DDTHH:MM:SS, or YYYY-MM-DDTHH:MM:SS+HH:MM."
            )
        suffix += f"&after={urllib.parse.quote(publish_after)}"
    publish_before = os.getenv("PUBLISH_BEFORE", None)
    if publish_before:
        if not is_valid_iso8601(publish_before):
            raise ValueError(
                f"Error: Invalid publish_before: {publish_before}. publish_before must be a valid ISO 8601 date string, in the format of YYYY-MM-DDTHH:MM:SS, or YYYY-MM-DDTHH:MM:SS+HH:MM."
            )
        suffix += f"&before={urllib.parse.quote(publish_before)}"

    modified_after = os.getenv("MODIFIED_AFTER", None)
    if modified_after:
        if not is_valid_iso8601(modified_after):
            raise ValueError(
                f"Error: Invalid modified_after: {modified_after}. modified_after must be a valid ISO 8601 date string, in the format of YYYY-MM-DDTHH:MM:SS, or YYYY-MM-DDTHH:MM:SS+HH:MM."
            )
        suffix += f"&modified_after={urllib.parse.quote(modified_after)}"

    modified_before = os.getenv("MODIFIED_BEFORE", None)
    if modified_before:
        if not is_valid_iso8601(modified_before):
            raise ValueError(
                f"Error: Invalid modified_before: {modified_before}. modified_before must be a valid ISO 8601 date string, in the format of YYYY-MM-DDTHH:MM:SS, or YYYY-MM-DDTHH:MM:SS+HH:MM."
            )
        suffix += f"&modified_before={urllib.parse.quote(modified_before)}"
    order = os.getenv("ORDER", "desc").lower()
    order_enum = ["asc", "desc"]
    if order not in order_enum:
        raise ValueError(
            f"Error: Invalid order: {order}. order must be one of: {order_enum}."
        )
    suffix += f"&order={order}"
    url += suffix

    response = client.get(url)
    if response.status_code >= 200 and response.status_code < 300:
        return [_format_posts_response(post) for post in response.json()]
    else:
        print(f"Failed to list posts. Error: {response.status_code}, {response.text}")


@tool_registry.register("CreatePost")
def create_post(client):

    url = f"{WORDPRESS_API_URL}/posts"

    title = os.getenv("TITLE", "")
    content = os.getenv("CONTENT", "")
    if title == "" and content == "":
        raise ValueError("Error: At least one of title or content must be provided.")
    status = os.getenv("STATUS", "publish").lower()
    status_enum = ["publish", "future", "draft", "pending", "private"]
    if status not in status_enum:
        raise ValueError(
            f"Error: Invalid status: {status}. status must be one of: {status_enum}."
        )
    comment_status = os.getenv("COMMENT_STATUS", "open").lower()
    comment_status_enum = ["open", "closed"]
    if comment_status not in comment_status_enum:
        raise ValueError(
            f"Error: Invalid comment_status: {comment_status}. comment_status must be one of: {comment_status_enum}."
        )
    sticky = str_to_bool(os.getenv("STICKY", "false"))

    post_data = {
        "title": title,
        "content": content,
        "status": status,
        "comment_status": comment_status,
        "sticky": sticky,
    }

    slug = os.getenv("SLUG", None)
    if slug:
        post_data["slug"] = slug
    date = os.getenv("DATE", None)
    if date:
        if not is_valid_iso8601(date):
            raise ValueError(
                f"Error: Invalid date: {date}. date must be a valid ISO 8601 date string, in the format of YYYY-MM-DDTHH:MM:SS, or YYYY-MM-DDTHH:MM:SS+HH:MM."
            )
        post_data["date"] = date

    format = os.getenv("FORMAT", "").lower()
    format_enum = [
        "standard",
        "aside",
        "chat",
        "gallery",
        "link",
        "image",
        "quote",
        "status",
        "video",
        "audio",
    ]
    if format != "":
        if format not in format_enum:
            raise ValueError(
                f"Error: Invalid format: {format}. format must be one of: {format_enum}."
            )
        post_data["format"] = format

    password = os.getenv("PASSWORD", None)
    if password:
        post_data["password"] = password

    author_id = os.getenv("AUTHOR_ID", None)
    if author_id:
        if not author_id.isdigit():
            raise ValueError(
                f"Error: Invalid author_id: {author_id}. author_id must be an integer."
            )
        post_data["author"] = int(author_id)
    excerpt = os.getenv("EXCERPT", None)
    if excerpt:
        post_data["excerpt"] = excerpt

    ping_status = os.getenv("PING_STATUS", "open").lower()
    ping_status_enum = ["open", "closed"]
    if ping_status not in ping_status_enum:
        raise ValueError(
            f"Error: Invalid ping_status: {ping_status}. ping_status must be one of: {ping_status_enum}."
        )
    post_data["ping_status"] = ping_status

    response = client.post(url, json=post_data)
    if response.status_code >= 200 and response.status_code < 300:
        return _format_posts_response(response.json())
    else:
        print(f"Failed to create post. Error: {response.status_code}, {response.text}")


@tool_registry.register("DeletePost")
def delete_post(client):
    post_id = os.environ["POST_ID"]
    force_delete = str_to_bool(os.getenv("FORCE_DELETE", "false"))
    url = f"{WORDPRESS_API_URL}/posts/{post_id}"
    if force_delete:
        url += "?force=true"

    response = client.delete(url)
    if response.status_code >= 200 and response.status_code < 300:
        return {"message": "Post deleted successfully"}
    else:
        print(f"Failed to delete post. Error: {response.status_code}, {response.text}")


@tool_registry.register("UpdatePost")
def update_post(client):
    post_id = os.environ["POST_ID"]
    url = f"{WORDPRESS_API_URL}/posts/{post_id}"

    post_data = {}

    if "TITLE" in os.environ:
        title = os.environ["TITLE"]
        if title == "":
            raise ValueError("Error: Title to update cannot be empty.")
        post_data["title"] = title
    if "CONTENT" in os.environ:
        content = os.environ["CONTENT"]
        if content == "":
            raise ValueError("Error: Content to update cannot be empty.")
        post_data["content"] = content
    if "STATUS" in os.environ:
        status = os.environ["STATUS"].lower()
        status_enum = ["publish", "future", "draft", "pending", "private"]
        if status not in status_enum:
            raise ValueError(
                f"Error: Invalid status: {status}. status must be one of: {status_enum}."
            )
        post_data["status"] = status
    if "COMMENT_STATUS" in os.environ:
        comment_status = os.environ["COMMENT_STATUS"].lower()
        comment_status_enum = ["open", "closed"]
        if comment_status not in comment_status_enum:
            raise ValueError(
                f"Error: Invalid comment_status: {comment_status}. comment_status must be one of: {comment_status_enum}."
            )
        post_data["comment_status"] = comment_status

    if "STICKY" in os.environ:
        sticky = str_to_bool(os.getenv("STICKY"))
        post_data["sticky"] = sticky

    if "SLUG" in os.environ:
        slug = os.environ["SLUG"]
        post_data["slug"] = slug
    if "DATE" in os.environ:
        date = os.environ["DATE"]
        if not is_valid_iso8601(date):
            raise ValueError(
                f"Error: Invalid date: {date}. date must be a valid ISO 8601 date string, in the format of YYYY-MM-DDTHH:MM:SS, or YYYY-MM-DDTHH:MM:SS+HH:MM."
            )
        post_data["date"] = date

    if "FORMAT" in os.environ:
        format = os.environ["FORMAT"].lower()
        format_enum = [
            "standard",
            "aside",
            "chat",
            "gallery",
            "link",
            "image",
            "quote",
            "status",
            "video",
            "audio",
        ]
        if format != "":
            if format not in format_enum:
                raise ValueError(
                    f"Error: Invalid format: {format}. format must be one of: {format_enum}."
                )
            post_data["format"] = format

    if "PASSWORD" in os.environ:
        password = os.environ["PASSWORD"]
        post_data["password"] = password

    if "AUTHOR_ID" in os.environ:
        author_id = os.environ["AUTHOR_ID"]
        if not author_id.isdigit():
            raise ValueError(
                f"Error: Invalid author_id: {author_id}. author_id must be an integer."
            )
        post_data["author"] = int(author_id)
    if "EXCERPT" in os.environ:
        excerpt = os.environ["EXCERPT"]
        post_data["excerpt"] = excerpt

    if "PING_STATUS" in os.environ:
        ping_status = os.environ["PING_STATUS"].lower()
        ping_status_enum = ["open", "closed"]
        if ping_status not in ping_status_enum:
            raise ValueError(
                f"Error: Invalid ping_status: {ping_status}. ping_status must be one of: {ping_status_enum}."
            )
        post_data["ping_status"] = ping_status

    response = client.post(url, json=post_data)
    if response.status_code >= 200 and response.status_code < 300:
        return _format_posts_response(response.json())
    else:
        print(f"Failed to update post. Error: {response.status_code}, {response.text}")
